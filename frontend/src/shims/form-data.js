/** Browser shim for Node `form-data` import in generated API client. */
const FD = typeof globalThis !== 'undefined' && globalThis.FormData
    ? globalThis.FormData
    : FormData;
export default FD;
