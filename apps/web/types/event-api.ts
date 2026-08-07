export type EventResponse = {
  name: string;
  namespace: string;
  reason: string;
  message: string;
  type: string;
  count: number;
  firstTimestamp?: string;
  lastTimestamp?: string;
  age: string;
  involvedObject: string;
  component: string;
  source: string;
};

export type EventListResponse = {
  items: EventResponse[];
  total: number;
};

export type EventQueryParams = {
  namespace?: string;
};
