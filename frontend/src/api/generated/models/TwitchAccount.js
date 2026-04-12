export var TwitchAccount;
(function (TwitchAccount) {
    let account_type;
    (function (account_type) {
        account_type["MAIN"] = "main";
        account_type["BOT"] = "bot";
    })(account_type = TwitchAccount.account_type || (TwitchAccount.account_type = {}));
})(TwitchAccount || (TwitchAccount = {}));
