/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { SystemStatsCaches } from './SystemStatsCaches';
import type { SystemStatsHost } from './SystemStatsHost';
import type { SystemStatsProcess } from './SystemStatsProcess';
import type { SystemStatsTables } from './SystemStatsTables';
export type SystemStatsResponse = {
    /**
     * When this snapshot was produced (server clock)
     */
    captured_at: string;
    tables: SystemStatsTables;
    process: SystemStatsProcess;
    host: SystemStatsHost;
    caches: SystemStatsCaches;
};

