<template>
    <div id="app">
        <div class="columns">
            <div class="column is-narrow">
                <Applications v-bind:apps="[{'id':1, 'title':'SelfBookingBusinessManager', 'services':[{'id':1, 'title':'Redis', 'type':'redis'}, {'id':2, 'title':'MySQL', 'type':'mysql'}]}]"/>
            </div>

            <div class="column">
              <router-view></router-view>
            </div>
        </div>
    </div>
</template>

<script>
    import Applications from './components/Applications.vue'

    export default {
        name: 'app',
        components: {
            Applications
        }
    }
</script>

<style>
    /*#app {*/
    /*  font-family: 'Avenir', Helvetica, Arial, sans-serif;*/
    /*  -webkit-font-smoothing: antialiased;*/
    /*  -moz-osx-font-smoothing: grayscale;*/
    /*  text-align: center;*/
    /*  color: #2c3e50;*/
    /*  margin-top: 60px;*/
    /*}*/
</style>
