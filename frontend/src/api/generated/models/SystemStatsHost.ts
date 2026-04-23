/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type SystemStatsHost = {
    /**
     * Host CPU utilization percent; null when unavailable
     */
    cpu_percent: number | null;
    memory_total_bytes: number;
    memory_used_bytes: number;
    memory_used_percent: number;
    /**
     * Path passed to disk usage (e.g. / or a drive root)
     */
    disk_path: string;
    disk_total_bytes: number;
    disk_used_bytes: number;
    disk_used_percent: number;
};

