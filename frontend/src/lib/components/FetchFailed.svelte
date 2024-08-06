<script lang="ts">
    import type {ApiErrorDto} from '$lib/api/open';

    export let error: ApiErrorDto | null | undefined;
</script>

<div class="error-container">
    <h4>Data could not be loaded.</h4>
    <p class="page-load-failed">{`${error?.status ?? '000'} - ${error?.statusText ?? 'unknown error'}`}</p>
    {#if error?.message}
        <p>{error?.message}</p>
    {/if}
</div>

<style>
    .error-container {
        display: flex;
        flex-flow: column;
        justify-content: center;
        gap:1rem;
        height: 100%;
    }

    .error-container .page-load-failed {
        font-size: 8cqw;
        font-weight: 600;
        letter-spacing: 0.1rem;
    }
</style>