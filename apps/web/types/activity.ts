export type ActivityStatus =
  | "success"
  | "warning"
  | "info"
  | "error";

export interface Activity {
  id: number;
  title: string;
  description: string;
  time: string;
  status: ActivityStatus;
}