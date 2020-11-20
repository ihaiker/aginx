<template>
    <Nodes include="true"/>
</template>

<script>
    import {Switch as cSwitch} from '@coreui/vue'
    import Nodes from "@/views/Nodes";

    export default {
        name: 'DefaultAside',
        components: {
            Nodes,
            cSwitch
        }
    }
</script>
