import type {
  QuotationProposal,
  QuotationProposalsResponse,
  QuotationRequest,
  QuotationRequestsResponse,
} from "@/shared/api/generated";
import { dateTime, exact, uuid } from "@/features/messaging/protected-bff";
export function validQuotationRequest(
  value: unknown,
): value is QuotationRequest {
  return (
    exact(value, [
      "id",
      "customerId",
      "title",
      "description",
      "categoryId",
      "localityId",
      "budgetMinor",
      "proposalDeadline",
      "state",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(value.id) &&
    uuid(value.customerId) &&
    typeof value.title === "string" &&
    typeof value.description === "string" &&
    uuid(value.categoryId) &&
    uuid(value.localityId) &&
    (value.budgetMinor === null ||
      (Number.isInteger(value.budgetMinor) && Number(value.budgetMinor) > 0)) &&
    dateTime(value.proposalDeadline) &&
    ["open", "accepted", "closed"].includes(String(value.state)) &&
    dateTime(value.createdAt) &&
    dateTime(value.updatedAt)
  );
}
export function validQuotationRequests(
  value: unknown,
): value is QuotationRequestsResponse {
  return (
    exact(value, ["requests"]) &&
    Array.isArray(value.requests) &&
    value.requests.every(validQuotationRequest)
  );
}
export function validQuotationProposal(
  value: unknown,
): value is QuotationProposal {
  return (
    exact(value, [
      "id",
      "requestId",
      "providerId",
      "priceMinor",
      "message",
      "availableAt",
      "estimatedMinutes",
      "expiresAt",
      "state",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(value.id) &&
    uuid(value.requestId) &&
    uuid(value.providerId) &&
    Number.isInteger(value.priceMinor) &&
    Number(value.priceMinor) > 0 &&
    typeof value.message === "string" &&
    dateTime(value.availableAt) &&
    (value.estimatedMinutes === null ||
      Number.isInteger(value.estimatedMinutes)) &&
    (value.expiresAt === null || dateTime(value.expiresAt)) &&
    ["submitted", "accepted", "rejected", "expired"].includes(
      String(value.state),
    ) &&
    dateTime(value.createdAt) &&
    dateTime(value.updatedAt)
  );
}
export function validQuotationProposals(
  value: unknown,
): value is QuotationProposalsResponse {
  return (
    exact(value, ["proposals"]) &&
    Array.isArray(value.proposals) &&
    value.proposals.every(validQuotationProposal)
  );
}
