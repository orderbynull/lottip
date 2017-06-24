const connStateStarted = 0xf4;
const connStateFinished = 0xf5;
const cmdResultError = 0xff;
const typingMessage = 'Waiting for you to stop typing...';
const copyDoneMessage = 'Copied to clipboard';
const executeUrl = '/execute';
const notificationShowTimeMs = 2000;

var ws;

new Vue({
    el: '#app',
    data: {
        connected: false,
        connections: {},
        backupConnections: null,
        connectionsStates: {},
        queriesCount: 0,
        filterQuery: '',
        tipMessage: '',
        modalQueryResult: ''
    },

    watch: {
        filterQuery: function () {
            this.tipMessage = typingMessage;
            this.getFilteredData();
        }
    },

    // Fired after app created
    created: function () {
        this.connect();
    },

    methods: {
        // Copies query string into clipboard
        copyQuery: function (connId, queryId) {
            if (clipboard.copy(this.connections[connId][queryId]['query']) && 'Notification' in window) {

                const notify = function () {
                    var notification = new Notification(copyDoneMessage, {requireInteraction: false});
                    setTimeout(notification.close.bind(notification), notificationShowTimeMs);
                };

                if (Notification.permission === 'granted') {
                    notify();
                }
                else if (Notification.permission !== 'denied') {
                    Notification.requestPermission(function (permission) {
                        if (permission === 'granted') {
                            notify();
                        }
                    });
                }
            }
        },

        // Sends query string to http endpoint and shows result in modal window
        executeQuery: function (connId, queryId) {
            if (this.connections[connId][queryId]['executable']) {
                var vue = this;

                vue.modalQueryResult = '';

                $('#results').modal();

                $.post(
                    executeUrl,
                    {
                        database: this.connections[connId][queryId]['database'],
                        query: this.connections[connId][queryId]['query']
                    },
                    function (data) {
                        vue.modalQueryResult = data;
                    }
                );
            }
        },

        // Filters queries by user provided string
        getFilteredData: _.debounce(function () {
            this.tipMessage = '';

            // Backup raw data if there's no backup yet
            if (this.backupConnections === null) {
                this.backupConnections = this.connections;
            }

            // Restore backup if filter is empty and backup exists
            if (this.filterQuery === '') {
                if (this.backupConnections !== null) {
                    this.connections = this.backupConnections;
                    this.backupConnections = null;
                }
                return;
            }

            var result = {};
            var connections = this.backupConnections !== null ? this.backupConnections : this.connections;

            for (conn in connections) {
                if (connections.hasOwnProperty(conn)) {

                    for (query in connections[conn]) {
                        if (connections[conn].hasOwnProperty(query)) {

                            if (connections[conn][query]['query'].toLowerCase().indexOf(this.filterQuery.toLowerCase()) >= 0) {
                                if (!(result[conn])) {
                                    result[conn] = {};
                                }
                                result[conn][query] = connections[conn][query];
                            }

                        }
                    }
                }
            }

            this.connections = result;
        }, 500),

        // Disconnects from websocket server
        disconnect: function () {
            this.connected && ws.close();
            console.error(this.connections);
        },

        // Connects to websocket server
        connect: function () {
            var app = this;

            // Connect back to the same addr this page was loaded from
            var parser = document.createElement('a');
            parser.href = window.location;
            ws = new WebSocket("ws://" + parser.host + "/ws");

            ws.onmessage = function (evt) {
                var data = JSON.parse(evt.data);

                //Cmd received
                if ('Query' in data) {
                    app.cmdReceived(data.ConnId, data.CmdId, data.Database, data.Query, data.Executable);
                    return;
                }

                //CmdResult received
                if ('Result' in data) {
                    app.cmdResultReceived(data.ConnId, data.CmdId, data.Result, data.Error, data.Duration);
                    return;
                }

                // ConnState received
                if ('State' in data) {
                    app.connStateReceived(data.ConnId, data.State);
                }
            };

            ws.onopen = function () {
                app.connected = true;
            };

            ws.onclose = function () {
                app.connected = false;
            };
        },

        // Returns if connection is still active or not
        isConnectionActive: function (connId) {
            return this.connectionsStates[connId] === connStateStarted;
        },

        // Clear all data to blank page
        clearAll: function () {
            this.connections = {};
            this.queriesCount = 0;
        },

        // Expand or collapse truncated query
        toggleExpandQuery: function (connId, cmdId) {
            this.connections[connId][cmdId].expanded =
                !this.connections[connId][cmdId].expanded;
        },

        // Fired when received Cmd data from websocket
        cmdReceived: function (connId, cmdId, database, query, executable) {
            if (!(connId in this.connections)) {
                Vue.set(this.connections, connId, {});
            }

            Vue.set(this.connections[connId], cmdId, {
                connId: connId,
                cmdId: cmdId,
                database: database,
                query: query,
                expanded: true,
                executable: executable,
                result: 'result-pending',
                duration: '?.??',
                error: ''
            });

            this.queriesCount++;

            Vue.set(this.connectionsStates, connId, connStateStarted);
        },

        // Fired when received CmdResult from websocket
        cmdResultReceived: function (connId, cmdId, result, error, duration) {
            if (this.connections[connId] !== undefined &&
                this.connections[connId][cmdId] !== undefined) {
                switch (result) {
                    case cmdResultError:
                        this.connections[connId][cmdId].result = 'result-error';
                        break;
                    default:
                        this.connections[connId][cmdId].result = 'result-ok';
                        break;
                }

                this.connections[connId][cmdId].duration = duration;
                this.connections[connId][cmdId].error = error;
            }
        },

        // Fired when received ConnState from websocket
        connStateReceived: function (connId, state) {
            Vue.set(this.connectionsStates, connId, state);
        }
    }
});