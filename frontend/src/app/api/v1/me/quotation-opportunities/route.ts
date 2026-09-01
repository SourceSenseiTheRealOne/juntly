import { listQuotationOpportunities } from "@/shared/api/generated";
import { authorizedHeaders, correlated, fail, requestID, requestIDHeader, sessionToken, unavailable } from "@/features/messaging/protected-bff";
import { validQuotationRequests } from "@/features/quotations/quotation-bff";
export const runtime="nodejs";
export async function GET(request:Request):Promise<Response>{const id=requestID(request.headers),token=await sessionToken();if(!token)return fail("UNAUTHORIZED","Unauthorized",401,id);const baseUrl=process.env.JUNTLY_API_ORIGIN;if(!baseUrl)return unavailable(id);try{const upstream=await listQuotationOpportunities({baseUrl,headers:authorizedHeaders(token,id)});if(upstream.error||!upstream.response?.ok||!correlated(upstream.response,id)||!validQuotationRequests(upstream.data))return unavailable(id);return Response.json(upstream.data,{status:200,headers:{[requestIDHeader]:id}})}catch{return unavailable(id)}}
