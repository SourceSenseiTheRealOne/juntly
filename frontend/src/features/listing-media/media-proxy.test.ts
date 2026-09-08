import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mediaProxy } from "./media-proxy";
import { GET as ownerList } from "@/app/api/v1/me/listings/[listingId]/media/route";
import { GET as moderatorList } from "@/app/api/v1/moderation/listings/[listingId]/media/route";
import { GET as publicList } from "@/app/api/v1/public/listings/[listingId]/media/route";
const mocks=vi.hoisted(()=>({token:vi.fn(),admin:vi.fn()}));
vi.mock("@/features/messaging/protected-bff",async(importOriginal)=>({...await importOriginal<object>(),sessionToken:mocks.token}));
vi.mock("@/features/auth/sole-administrator",()=>({resolveSoleAdministratorSession:mocks.admin}));
const listingId="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const mediaId="bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
const requestId="req_media_proxy_test";
function request(method="GET"){return new Request("http://localhost:4200/api/v1/public/listings/"+listingId+"/media",{method,headers:{"X-Request-ID":requestId}})}
function upstream(body:unknown,status=200){return Response.json(body,{status,headers:{"X-Request-ID":requestId}})}
describe("media BFF boundary",()=>{
  beforeEach(()=>{vi.stubEnv("JUNTLY_API_ORIGIN","http://127.0.0.1:9000");mocks.token.mockResolvedValue(null);mocks.admin.mockResolvedValue({status:"forbidden"});vi.stubGlobal("fetch",vi.fn());});
  afterEach(()=>{vi.unstubAllGlobals();vi.unstubAllEnvs();vi.clearAllMocks();});
  it("rejects owner and non-sole-admin reads before upstream access",async()=>{
    expect((await ownerList(request(),{params:Promise.resolve({listingId})})).status).toBe(401);
    expect((await moderatorList(request(),{params:Promise.resolve({listingId})})).status).toBe(403);
    expect(fetch).not.toHaveBeenCalled();
  });
  it("returns only correlated bounded metadata without forwarding browser credentials",async()=>{
    vi.mocked(fetch).mockResolvedValue(upstream({photos:[]}));
    const response=await publicList(request(),{params:Promise.resolve({listingId})});
    expect(response.status).toBe(200);expect(await response.json()).toEqual({photos:[]});
    const [,options]=vi.mocked(fetch).mock.calls[0];
    expect(options?.cache).toBe("no-store");expect(options?.redirect).toBe("error");
    expect(new Headers(options?.headers).has("Authorization")).toBe(false);
    vi.mocked(fetch).mockResolvedValue(upstream({photos:[],objectReference:"secret"}));
    expect((await mediaProxy(request(),{listingId},"public","list")).status).toBe(503);
  });
  it("preserves authenticated 204 finalization and rejects cross-origin writes",async()=>{
    mocks.token.mockResolvedValue("synthetic-session-token");
    vi.mocked(fetch).mockResolvedValue(new Response(null,{status:204,headers:{"X-Request-ID":requestId}}));
    const response=await mediaProxy(request("POST"),{listingId,mediaId},"me","finalize");
    expect(response.status).toBe(204);
    expect(await response.text()).toBe("");
    const r=request("POST");r.headers.set("Origin","https://attacker.example");
    expect((await mediaProxy(r,{listingId,mediaId},"me","finalize")).status).toBe(403);
    expect(fetch).toHaveBeenCalledTimes(1);
  });
  it("requires PNG content, bounds binary responses and never proxies provider errors",async()=>{
    vi.mocked(fetch).mockResolvedValue(new Response("not PNG",{headers:{"Content-Type":"text/html","X-Request-ID":requestId}}));
    expect((await mediaProxy(request(),{listingId,mediaId},"public","image")).status).toBe(503);
    vi.mocked(fetch).mockResolvedValue(upstream({error:{code:"NOT_FOUND",message:"private detail",requestId}},404));
    const response=await mediaProxy(request(),{listingId,mediaId},"public","image");
    expect(response.status).toBe(404);expect(await response.text()).not.toContain("private detail");
    vi.mocked(fetch).mockResolvedValue(new Response(new Uint8Array(10485761),{headers:{"Content-Type":"image/png","X-Request-ID":requestId}}));
    expect((await mediaProxy(request(),{listingId,mediaId},"public","image")).status).toBe(503);
  });
});
