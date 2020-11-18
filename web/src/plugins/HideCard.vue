<template>
    <div class="card card-accent-primary">
        <div class="card-header" @click="showBody = !showBody">{{ title }}
            <div class="card-header-actions">
                <div class="card-header-action btn-minimize">
                    <i :class="showBody?'icon-arrow-up':'icon-arrow-down'"></i>
                </div>
            </div>
        </div>

        <div class="collapse" :class="showBody?'show':'hide'">
            <div class="card-body p-2">
                <slot></slot>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "HideCard",
    data: () => ({showBody: false }),
    props: {
        title: {
            type: String
        },
        show : Boolean,
    },
    mounted() {
        this.showBody = this.show === true;
    }
}
</script>
