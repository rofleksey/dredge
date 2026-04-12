import { ApiError, ClientNotice } from '../api/generated';
import { notify } from './notify';
function parseClientNoticeBody(body) {
    if (!body || typeof body !== 'object') {
        return null;
    }
    const o = body;
    const sev = o.severity;
    const msg = o.message;
    if (typeof msg !== 'string') {
        return null;
    }
    if (sev === ClientNotice.severity.WARNING) {
        return { severity: 'warning', message: msg };
    }
    if (sev === ClientNotice.severity.ERROR) {
        return { severity: 'error', message: msg };
    }
    return null;
}
/** Maps API errors with ClientNotice JSON body to toast notifications. */
export function notifyFromApiError(err, opts) {
    if (err instanceof ApiError) {
        const n = parseClientNoticeBody(err.body);
        if (n) {
            notify({
                id: opts.id,
                type: n.severity === 'warning' ? 'warning' : 'error',
                title: opts.title,
                description: n.message,
            });
            return;
        }
    }
    notify({
        id: opts.id,
        type: 'error',
        title: opts.title,
        description: opts.fallbackDescription,
    });
}
