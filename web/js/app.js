const connStateStarted = 0xf4;
const connStateFinished = 0xf5;
const cmdResultError = 0xff;
const typingMessage = 'Waiting for you to stop typing...';

var ws;

new Vue({
    el: '#app',
    data: {
        connected: false,
        connections: {},
        backupConnections: null,
        connectionsStates: {},
        queriesCount: 0,
        globalExpand: false,
        filterQuery: '',
        tipMessage: ''
    },

    watch: {
        filterQuery: function () {
            console.log(this.filterQuery);
            this.tipMessage = typingMessage;
            this.getFilteredData();
        }
    },

    created: function () {
        this.connect();
    },

    methods: {
        getFilteredData: _.debounce(function () {
            this.tipMessage = '';

            if (this.backupConnections === null) {
                this.backupConnections = this.connections;
            }

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

        disconnect: function () {
            this.connected && ws.close();
            console.error(this.connections);
        },

        connect: function () {
            var app = this;

            var parser = document.createElement('a');
            parser.href = window.location;

            ws = new WebSocket("ws://" + parser.host + "/ws");

            ws.onmessage = function (evt) {
                var data = JSON.parse(evt.data);

                //Cmd received
                if ('Query' in data) {
                    app.cmdReceived(data.ConnId, data.CmdId, data.Query);
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

        // Globally expand or collapse all queries
        toggleGlobalExpand: function () {
            this.globalExpand = !this.globalExpand;
        },

        // Expand or collapse truncated query
        toggleExpandQuery: function (connId, cmdId) {
            this.connections[connId][cmdId].expanded =
                !this.connections[connId][cmdId].expanded;
        },

        // Fired when received Cmd data from websocket
        cmdReceived: function (connId, cmdId, query) {
            if (!(connId in this.connections)) {
                Vue.set(this.connections, connId, {});
            }

            Vue.set(this.connections[connId], cmdId, {
                connId: connId,
                cmdId: cmdId,
                expanded: false,
                query: query,
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