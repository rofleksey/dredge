export var ClientNotice;
(function (ClientNotice) {
    let severity;
    (function (severity) {
        severity["ERROR"] = "error";
        severity["WARNING"] = "warning";
    })(severity = ClientNotice.severity || (ClientNotice.severity = {}));
})(ClientNotice || (ClientNotice = {}));
