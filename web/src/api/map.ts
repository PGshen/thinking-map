import { post } from "./request";
import API_ENDPOINTS from "./endpoints";
import type { CreateMapRequest, CreateMapResponse } from "@/types/map";

export async function createMap(params: CreateMapRequest): Promise<CreateMapResponse> {
  return post(API_ENDPOINTS.MAP.CREATE, params);
} 