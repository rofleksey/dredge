/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ClientNotice = {
    severity: ClientNotice.severity;
    code: string;
    message: string;
    details?: Record<string, any>;
};
export namespace ClientNotice {
    export enum severity {
        ERROR = 'error',
        WARNING = 'warning',
    }
}

