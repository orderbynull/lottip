<template>
    <aside class="menu">
        <p class="menu-label">
            Applications
        </p>
        <ul class="menu-list">
            <li>
                <a class="is-active">SelfBookingBusinessManager</a>
                <ul>
                    <li><router-link to="/mysql">MySQL</router-link></li>
                    <li><router-link to="/redis">Redis</router-link></li>
                    <li><a>RabbitMQ</a></li>
                </ul>
            </li>
            <li><a>Scoring</a></li>
            <li><a>PaymentLinks</a></li>
            <li><a>Autofunnel</a></li>
        </ul>
    </aside>
</template>

<script>
    export default {
        name: 'Applications',
        props: {
            msg: String
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