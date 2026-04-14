export type Nullable<T> = T | null

export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string }

export type PaginatedResponse<T> = {
  items: T[]
  total: number
  limit: number
  offset: number
}
