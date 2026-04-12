export var NotificationEntry;
(function (NotificationEntry) {
    let provider;
    (function (provider) {
        provider["TELEGRAM"] = "telegram";
        provider["WEBHOOK"] = "webhook";
    })(provider = NotificationEntry.provider || (NotificationEntry.provider = {}));
})(NotificationEntry || (NotificationEntry = {}));
