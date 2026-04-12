export var ChatHistoryEntry;
(function (ChatHistoryEntry) {
    /**
     * irc: observed in chat; sent: posted via dredge
     */
    let source;
    (function (source) {
        source["IRC"] = "irc";
        source["SENT"] = "sent";
    })(source = ChatHistoryEntry.source || (ChatHistoryEntry.source = {}));
})(ChatHistoryEntry || (ChatHistoryEntry = {}));
