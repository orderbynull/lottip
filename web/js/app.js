var ws = null;

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
            showSession: 0,
            items: [],
            sessions: [],
        },
        methods: {
            setSession: function (session) {
                this.showSession = session
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
            },

            showDetails: function (index) {
                //
                //this.items[index].detailed = !this.items[index].detailed;
            },

            clearItems: function () {
                this.items = [];
                this.sessions = [];
                this.showSession = 0;
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
                                time: new Date().toLocaleTimeString(),
                                query: data.Query,
                                detailed: vm.isExpanded,
                                sessId: data.SessionID
                            });

                            if (vue.showSession === 0) {
                                vue.showSession = data.SessionID;
                            }


                            if (!vue.sessions.some(function (e) { return e.id == data.SessionID })) {
                                vue.sessions.push({ id: data.SessionID, inProgress: true });
                            }

                            break;

                        case "State":
                        console.log(data);
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