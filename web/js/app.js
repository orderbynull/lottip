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

                if(this.isExpanded){
                    $('pre code').each(function(i, block) {
                        hljs.highlightBlock(block);
                    });
                }
            },

            showDetails: function (id) {
                this.items[id].detailed = !this.items[id].detailed;
                hljs.highlightBlock($('pre#snip-'+id+' code').get(0));
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
                    console.log('CONNECTED');
                    vue.isConnected = true;
                }

                ws.onclose = function (evt, reason) {
                    console.log('DISCONNECTED ' + reason);
                    vue.isConnected = false;
                    ws = null;
                }

                ws.onmessage = function (evt) {
                    data = jQuery.parseJSON(evt.data);

                    //console.log(data.Type);

                    switch (data.Type) {
                        case "Query":
                            vue.items.push({
                                id: queryID++,
                                time: new Date().toLocaleTimeString(),
                                query: data.Query,
                                detailed: vm.isExpanded,
                                sessId: data.SessionID
                            });

                            //console.log(vue.items);

                            if (vue.activeSession === 0) {
                                vue.activeSession = data.SessionID;
                            }

                            if (!vue.sessions.some(function (e) { return e.id == data.SessionID })) {
                                vue.sessions.push({ id: data.SessionID, inProgress: true });
                            }

                            break;

                        case "State":
                            for (var i in vue.sessions) {
                                if (vue.sessions[i].id == data.SessionID) {
                                    vue.sessions[i].inProgress = data.State;
                                }
                            }
                            break;
                    }
                }

                ws.onerror = function (evt) {
                    console.log("ERROR: " + evt.data);
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