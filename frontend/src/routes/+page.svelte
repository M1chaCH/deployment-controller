<script lang="ts">
    import {isErrorDto} from '$lib/api/open';
    import {pagesStore} from '$lib/api/store';
    import PageOutline from '$lib/PageOutline.svelte';
</script>

<PageOutline pageName="Home">
    <div slot="description">
        <p>This page shows all applications that are deployed on my server. Feel free to explore and get to know my projects more closely.</p>
        <p class="subtext">Some might be locked, you need to have special access for these.</p>
    </div>
    <div slot="content">
        <h3>Deployments</h3>
        {#if $pagesStore !== null && !isErrorDto($pagesStore) }
            <div class="page-container">
                {#each $pagesStore as page }
                    <div class="page-card">
                        <h4>{page.pageTitle}</h4>
                        <div class="page-content">
                            <p>{page.pageDescription}</p>
                        </div>
                        <a class="page-link" class:disabled={!page.accessAllowed} href={page.accessAllowed ? page.pageUrl : ""}>
                            <span class="material-symbols-outlined">jump_to_element</span>
                        </a>
                        {#if !page.accessAllowed }
                            <p class="page-no-access">Locked</p>
                        {/if}
                    </div>
                {/each}
            </div>
        {:else}
            <div class="error-container">
                <h4>Pages could not be loaded.</h4>
                <p class="page-load-failed">{`${$pagesStore?.status ?? '000'} - ${$pagesStore?.statusText ?? 'unknown error'}`}</p>
                {#if $pagesStore?.message}
                    <p>{$pagesStore.message}</p>
                {/if}
            </div>
        {/if}
    </div>
</PageOutline>

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

    .page-container {
        display: flex;
        flex-flow: row wrap;
        gap: 2rem;
    }

    .page-card {
        flex: 1 1 30%;
        min-width: 300px;
        max-width: 780px;
        min-height: 360px;

        box-sizing: border-box;

        display: flex;
        flex-flow: column;
    }

    .page-card h4 {
        margin: 1rem 0 0.2rem 1rem;
    }

    .page-content {
        border-left: 2px solid var(--controller-line-color);
        background-color: var(--controller-area-color);
        flex: 4;
        padding: 1rem;
        box-sizing: border-box;
    }

    .page-link {
        background-color: var(--michu-tech-primary);
        padding-top: 0.5rem;
        padding-right: 1rem;
        padding-bottom: 1rem;
        box-sizing: border-box;

        display:flex;
        flex-flow: row nowrap;
        justify-content: flex-end;

        position: relative;
    }

    .page-link::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        width: 0;
        background-color: var(--michu-tech-accent);
        transition: width 500ms ease-in-out;
        z-index: 1;
    }

    .page-link:hover:not(.disabled)::before {
        width: 100%;
    }

    .page-link span {
        color: var(--michu-tech-white);
        font-size: 1.6rem;
        z-index: 2;
    }

    .page-link.disabled {
        opacity: 0.5;
        cursor: default;
    }

    .page-no-access {
        background-color: var(--michu-tech-warn);
        color: var(--michu-tech-white);
        font-size: 0.8rem;
        font-weight: 600;
        text-align: center;
        padding: 0.1rem;
    }
</style>
