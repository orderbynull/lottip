var ws = null;
var queryID = 0;

$(document).ready(function () {
    if (!window.WebSocket) alert("WebSocket not supported by this browser");

    Vue.component('demo-grid', {
        template: '#grid-template',
        props: {
            data: Array,
            columns: Array,
            filterKey: String
        },
        data: function () {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
            return {
                sortKey: '',
                sortOrders: sortOrders
            }
        }
    })

    Vue.filter('exactFilterBy', function (array, needle, inKeyword, key) {
        return array.filter(function (item) {
            return item[key] == needle;
        });
    });

    var app = new Vue({
        el: '#app',
        data: {
            earchQuery: '',
            gridColumns: ['name', 'power'],
            gridData: [
                { name: 'Chuck Norris', power: Infinity },
                { name: 'Bruce Lee', power: 9000 },
                { name: 'Jackie Chan', power: 7000 },
                { name: 'Jet Li', power: 8000 }
            ],

            showId: false,
            showTime: false,
            showSidebar: true,
            isExpanded: false,
            isConnected: false,
            activeSession: 0,
            activeSessionIndex: 1,
            items: [],
            sessions: [],
        },
        methods: {
            toggleSidebar: function () {
                this.showSidebar = !this.showSidebar;
            },

            executeSql: function (id) {
                this.showResults(id);
            },

            showSql: function (id) {
                this.items[id].detailed = true;
                this.items[id].showSql = true;
                this.items[id].showResults = false;
            },

            showResults: function (id) {
                this.items[id].detailed = true;
                this.items[id].showSql = false;
                this.items[id].showResults = true;
            },

            getSessionIndex: function (sessionID) {
                var key = null;

                this.sessions.every(function (session, index) {
                    if (session.id === sessionID) {
                        key = index;
                        return false;
                    }
                    return true;
                });

                return key;
            },

            /**
             * Handle data that holds query came from WebSocket connection
             */
            handleQuery: function (json) {
                if (!this.isConnected) {
                    return;
                }

                this.items.push({
                    id: queryID++,
                    time: new Date().toLocaleTimeString(),
                    query: data.Query,
                    detailed: this.isExpanded,
                    sessId: data.SessionID,
                    actions: false,
                    showSql: true,
                    showResults: false
                });

                if (this.activeSession === 0) {
                    this.activeSession = data.SessionID;
                }

                var index = this.getSessionIndex(data.SessionID);

                if (index === null) {
                    this.sessions.push({
                        id: data.SessionID,
                        inProgress: true,
                        queries: 1
                    });
                }
                else {
                    this.sessions[index]['queries']++;
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
            },

            /**
             * Toggle single query mode - expanded or collapsed
             */
            showDetails: function (id) {
                this.items[id].detailed = !this.items[id].detailed;
            },

            /**
             * Toggle query actions buttons
             */
            toggleQueryActions: function (id) {
                this.items[id].actions = !this.items[id].actions;
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

    app.startWs();
});