export var UserActivityEvent;
(function (UserActivityEvent) {
    let event_type;
    (function (event_type) {
        event_type["CHAT_ONLINE"] = "chat_online";
        event_type["CHAT_OFFLINE"] = "chat_offline";
        event_type["MESSAGE"] = "message";
    })(event_type = UserActivityEvent.event_type || (UserActivityEvent.event_type = {}));
})(UserActivityEvent || (UserActivityEvent = {}));
