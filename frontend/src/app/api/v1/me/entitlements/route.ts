import { getMyEntitlements } from "@/shared/api/generated";
import { authorizedHeaders, correlated, fail, requestID, requestIDHeader, sessionToken, unavailable } from "@/features/messaging/protected-bff";
import { validEntitlements } from "@/features/entitlements/entitlement-bff";
export const runtime="nodejs";
export async function GET(request:Request):Promise<Response>{const id=requestID(request.headers),token=await sessionToken();if(!token)return fail("UNAUTHORIZED","Unauthorized",401,id);const baseUrl=process.env.JUNTLY_API_ORIGIN;if(!baseUrl)return unavailable(id);try{const up=await getMyEntitlements({baseUrl,headers:authorizedHeaders(token,id)});if(up.error||!up.response?.ok||!correlated(up.response,id)||!validEntitlements(up.data))return unavailable(id);return Response.json(up.data,{status:200,headers:{[requestIDHeader]:id}})}catch{return unavailable(id)}}
