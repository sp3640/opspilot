export const PAGINATION_DEFAULT_PAGE = 1;
export const PAGINATION_DEFAULT_PAGE_SIZE = 6;
export const PAGINATION_MAX_PAGE_SIZE = 100;

export const SORT_ORDERS = {
  ASC: "asc",
  DESC: "desc",
} as const;

export const PROJECT_SORT_FIELDS = {
  NAME: "name",
  UPDATED_AT: "updated_at",
} as const;

export const PROJECT_LOOKUP_QUERY = {
  page: PAGINATION_DEFAULT_PAGE,
  limit: PAGINATION_MAX_PAGE_SIZE,
  sort: PROJECT_SORT_FIELDS.NAME,
  order: SORT_ORDERS.ASC,
} as const;
