<template>
    <nav aria-label="navigation">
        <ul class="pagination pagination-sm justify-content-center mb-0">
            <li class="page-item disabled">
                <a class="page-link" href="#" tabindex="-1" aria-disabled="true">
                    共 {{ total }} 条
                </a>
            </li>
            <li class="page-item" :class="{'disabled':start}">
                <button class="page-link" @click="toPage(1)">首页</button>
            </li>

            <li v-for="(p) in pages" class="page-item" :class="{'active':(p === page)}">
                <button class="page-link" @click="toPage(p)">{{ p }}</button>
            </li>

            <li class="page-item" :class="{'disabled':end}">
                <button class="page-link" @click="toPage(pages.length)">尾页</button>
            </li>

            <li class="page-item disabled">
                <a class="page-link" href="#" tabindex="-1" aria-disabled="true">
                    每页 {{ limit }} 条
                </a>
            </li>
        </ul>
    </nav>
</template>
<script>
export default {
    name: "XPage",
    props: {items: Object},
    methods: {
        toPage(page) {
            this.$emit("change", page);
        },
    },
    computed: {
        start() {
            return this.page === 1 || this.total < this.limit;
        },
        end() {
            return this.page === this.pages.length || this.total < this.limit;
        },
        total() {
            return this.items.total;
        },
        page() {
            return this.items.page;
        },
        limit() {
            return this.items.limit;
        },
        pages() {
            let pages = [];
            console.log("total:", this.total, ", limit:", this.limit,
                " pages: ", Math.ceil(this.total / this.limit))

            for (let i = 1; i <= Math.ceil(this.total / this.limit); i++) {
                pages.push(i);
            }
            return pages;
        }
    }
}
</script>
