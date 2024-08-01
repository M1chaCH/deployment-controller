<script lang="ts">

    export let title: string;
</script>

<div class="app-card">
    <h4>{title}</h4>
    <div class="card-content">
        <slot name="content"></slot>
    </div>
    <slot name="footer"></slot>
</div>

<style>
    .app-card {
        flex: 1 1 30%;
        min-width: 300px;
        max-width: 780px;
        min-height: 360px;
        height: 100%;

        box-sizing: border-box;

        display: flex;
        flex-flow: column;
    }

    .app-card h4 {
        margin: 1rem 0 0.2rem 1rem;
    }

    .card-content {
        border-left: 2px solid var(--controller-line-color);
        background-color: var(--controller-area-color);
        flex: 4;
        padding: 1rem;
        box-sizing: border-box;
    }
</style>
