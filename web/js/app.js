var ws = null;

$(document).ready(function () {

    var vm = new Vue({
        el: '#app',
        data: {
            showId: false,
            showTime: false,
            isExpanded: true,
            isConnected: false,
            items: []
        },
        methods: {
            toggleShowId: function () {
                this.showId = !this.showId;
            },

            toggleShowTime: function () {
                this.showTime = !this.showTime;
            },

            toggleExpanded: function () {
                var v = this;
                this.isExpanded = !this.isExpanded;
                this.items.forEach(function(item, index){
                    item.detailed = v.isExpanded;
                });
            },

            showDetails: function (index) {
                this.items[index].detailed = !this.items[index].detailed;
            },

            clearItems: function () {
                this.items = [];
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
                    console.log('DISCONNECTED '+reason);
                    vue.isConnected = false;
                    ws = null;
                }

                ws.onmessage = function (evt) {
                    data = jQuery.parseJSON(evt.data);

                    vue.items.push({time: new Date().toLocaleTimeString(), query: data.Query, detailed: vm.isExpanded});
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
    
//    vm.items.push({time: new Date().toLocaleTimeString('en-US', { hour12: false }), query: "SELECT self.*,self.name_ru as name, self.map_img_ext_ru as map_img_ext, resort.name_ru as resort_name, resort.name_en as resort_name_en, resort.country as country, country.name_ru as country_name, country.name_en as country_name_en, country.name_en as country_name_eng, country.region as region, country.sng as sng, country.sng as is_sng, country.iso_code as country_iso_code FROM uts_city as self LEFT JOIN uts_resort resort ON resort.id = self.resort LEFT JOIN uts_country country ON country.id = resort.country WHERE self.id = '4' AND self.name_ru<>'' AND self.name_ru IS NOT NULL AND self.name_en<>'' AND self.name_en IS NOT NULL LIMIT 1", detailed: false});

});