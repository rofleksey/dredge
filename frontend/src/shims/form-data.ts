/** Browser shim for Node `form-data` import in generated API client. */
const FD =
  typeof globalThis !== 'undefined' && (globalThis as unknown as { FormData?: typeof FormData }).FormData
    ? (globalThis as unknown as { FormData: typeof FormData }).FormData
    : FormData;
export default FD;
