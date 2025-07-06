import { post, get } from "./request";
import API_ENDPOINTS from "./endpoints";
import type { CreateMapRequest, CreateMapResponse, MapListResponse, MapListQuery } from "@/types/map";

export async function createMap(params: CreateMapRequest): Promise<CreateMapResponse> {
  return post(API_ENDPOINTS.MAP.CREATE, params);
}

export async function fetchMapList(params: MapListQuery): Promise<MapListResponse> {
  // Compose query string
  const query = new URLSearchParams();
  if (params.page) query.append("page", String(params.page));
  if (params.limit) query.append("limit", String(params.limit));
  if (typeof params.status === "number") query.append("status", String(params.status));
  if (params.problemType) query.append("problemType", params.problemType);
  if (params.dateRange) query.append("dateRange", params.dateRange);
  if (params.search) query.append("search", params.search);
  const url = `/v1/maps${query.toString() ? `?${query.toString()}` : ""}`;
  return get(url);
} 