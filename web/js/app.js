var ws = null;
var queryID = 0;

$(document).ready(function () {
    if (!window.WebSocket) alert("WebSocket not supported by this browser");

    Vue.filter('exactFilterBy', function (array, needle, inKeyword, key) {
        return array.filter(function (item) {
            return item[key] == needle;
        });
    });

    new Vue({
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
            highlightVisible: function () {
                $('pre code').each(function (i, block) {
                    hljs.highlightBlock(block);
                });
            },

            highlightOne: function (id) {
                hljs.highlightBlock($('pre#snip-' + id + ' code').get(0));
            },

            /**
             * Handle data that holds query came from WebSocket connection
             */
            handleQuery: function (json) {
                if(!this.isConnected){
                    return;
                }

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
                    this.highlightVisible();
                }
            },

            /**
             * Handle data that holds session state(active/done) came from WebSocket connection
             */
            handleSession: function (json) {
                for (var i in this.sessions) {
                    if (this.sessions[i].id == data.SessionID) {
                        this.sessions[i].inProgress = data.State;
                    }
                }
            },

            /**
             * Set currently active session that should be shown to user
             */
            setSession: function (sessionID, sessionIndex) {
                this.activeSession = sessionID;
                this.activeSessionIndex = sessionIndex;

                if (this.isExpanded) {
                    this.highlightVisible()
                }
            },

            /**
             * Toggle visibility of query id
             */
            toggleShowId: function () {
                this.showId = !this.showId;
            },

            /**
             * Toggle visibility of query datetime
             */
            toggleShowTime: function () {
                this.showTime = !this.showTime;
            },

            /**
             * Toggle all queries mode - expanded or collapsed
             */
            toggleExpanded: function () {
                var v = this;
                this.isExpanded = !this.isExpanded;
                this.items.forEach(function (item, index) {
                    item.detailed = v.isExpanded;
                });

                if (this.isExpanded) {
                    this.highlightVisible()
                }
            },

            /**
             * Toggle single query mode - expanded or collapsed
             */
            showDetails: function (id) {
                this.items[id].detailed = !this.items[id].detailed;
                this.highlightOne(id);
            },

            /**
             * Remove all queries and sessions
             */
            clearItems: function () {
                this.items = [];
                this.sessions = [];
                this.activeSession = 0;
                this.activeSessionIndex = 1;
                queryID = 0;
            },

            /**
             * Init WebSocket connection
             */
            startWs: function () {
                var vue = this;

                ws = new WebSocket("ws://127.0.0.1:8080/proxy");

                ws.onopen = function (evt) {
                    vue.isConnected = true;
                }

                ws.onclose = function (evt, reason) {
                    vue.isConnected = false;
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

            /**
             * Close WebSocket connection
             */
            stopWs: function () {
                if (ws !== null) {
                    ws.close();
                }
            }
        }
    });
});