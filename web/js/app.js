var ws = null;
var queryID = 0;

$(document).ready(function () {

    Vue.filter('exactFilterBy', function (array, needle, inKeyword, key) {
        return array.filter(function (item) {
            return item[key] == needle;
        });
    });

    var vm = new Vue({
        el: '#app',
        data: {
            showId: false,
            showTime: false,
            isExpanded: false,
            isConnected: false,
            activeSession: 0,
            activeSessionIndex: 1,
            items: [],
            sessions: [],
        },
        methods: {
            handleQuery: function (json) {
                this.items.push({
                    id: queryID++,
                    time: new Date().toLocaleTimeString(),
                    query: data.Query,
                    detailed: this.isExpanded,
                    sessId: data.SessionID
                });

                if (this.activeSession === 0) {
                    this.activeSession = data.SessionID;
                }

                if (!this.sessions.some(function (e) { return e.id == data.SessionID })) {
                    this.sessions.push({ id: data.SessionID, inProgress: true });
                }

                if (this.isExpanded) {
                    $('pre code').each(function (i, block) {
                        hljs.highlightBlock(block);
                    });
                }
            },

            handleSession: function (json) {
                for (var i in this.sessions) {
                    if (this.sessions[i].id == data.SessionID) {
                        this.sessions[i].inProgress = data.State;
                    }
                }
            },

            setSession: function (sessionID, sessionIndex) {
                this.activeSession = sessionID;
                this.activeSessionIndex = sessionIndex;
            },

            toggleShowId: function () {
                this.showId = !this.showId;
            },

            toggleShowTime: function () {
                this.showTime = !this.showTime;
            },

            toggleExpanded: function () {
                var v = this;
                this.isExpanded = !this.isExpanded;
                this.items.forEach(function (item, index) {
                    item.detailed = v.isExpanded;
                });

                if (this.isExpanded) {
                    $('pre code').each(function (i, block) {
                        hljs.highlightBlock(block);
                    });
                }
            },

            showDetails: function (id) {
                this.items[id].detailed = !this.items[id].detailed;
                hljs.highlightBlock($('pre#snip-' + id + ' code').get(0));
            },

            clearItems: function () {
                this.items = [];
                this.sessions = [];
                this.activeSession = 0;
                this.activeSessionIndex = 1;
                queryID = 0;
            },

            startWs: function () {
                if (ws !== null)
                    return;

                var vue = this;

                ws = new WebSocket("ws://127.0.0.1:8080/proxy");

                ws.onopen = function (evt) {
                    vue.isConnected = true;
                }

                ws.onclose = function (evt, reason) {
                    vue.isConnected = false;
                    ws = null;
                }

                ws.onmessage = function (evt) {
                    data = jQuery.parseJSON(evt.data);

                    switch (data.Type) {
                        case "Query":
                            vue.handleQuery(data);
                            break;

                        case "State":
                            vue.handleSession(data);
                            break;
                    }
                }
            },

            stopWs: function () {
                if (ws !== null)
                    ws.close();

                ws = null;

                this.isConnected = false;
            }
        }
    });
});