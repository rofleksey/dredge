export var CreateNotificationRequest;
(function (CreateNotificationRequest) {
    let provider;
    (function (provider) {
        provider["TELEGRAM"] = "telegram";
        provider["WEBHOOK"] = "webhook";
    })(provider = CreateNotificationRequest.provider || (CreateNotificationRequest.provider = {}));
})(CreateNotificationRequest || (CreateNotificationRequest = {}));
