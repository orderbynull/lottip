<template>
    <aside class="menu">
        <p class="menu-label">
            Applications
        </p>
        <ul class="menu-list">
            <li v-for="app in apps" v-bind:key="app.id">
                <a>{{app.title}}</a>
                <ul>
                    <li v-for="service in app.services" v-bind:key="service.id">
                        <router-link :to="service.type">{{service.title}}</router-link>
                    </li>
                </ul>
            </li>
        </ul>
    </aside>
</template>

<script>
    export default {
        name: 'Applications',
        props: {
            apps: Array
        }
    }
</script>

<style scoped>
    .sidebar {
        position: fixed;
        top: 0;
        bottom: 0;
        left: 0;
        z-index: 100;
        padding: 48px 0 0;
        box-shadow: inset -1px 0 0 rgba(0, 0, 0, .1);
    }
</style>