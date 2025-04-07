<script lang="ts">
    import {isErrorDto} from '$lib/api/open';
    import {pagesStore, userStore} from '$lib/api/store';
    import AppCard from '$lib/components/AppCard.svelte';
    import FetchFailed from '$lib/components/FetchFailed.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';

    // TODO implement change MFA type
</script>

<PageOutline pageName="Home">
    <div slot="description">
        <p>This page shows all applications that are deployed on my server. Feel free to explore and get to know my projects more closely.</p>
        <p class="subtext">Some might be locked, you need to have special access for these.</p>
        {#if $userStore && !isErrorDto($userStore) && !$userStore.onboard }
            <p style="margin-top: 2rem;">Your are not yet onboard. Make sure to <a href="/onboarding#onboarding" style="text-decoration: underline; color: var(--michu-tech-accent);">complete your setup</a> to access private pages.</p>
        {/if}
    </div>
    <div slot="content">
        <h3>Deployments</h3>
        {#if $pagesStore !== null && !isErrorDto($pagesStore) }
            <div class="page-container">
                {#each $pagesStore as page }
                    <AppCard title={page.pageTitle}>
                        <p slot="content">{page.pageDescription}</p>
                        <div slot="footer">
                            <a class="page-link" class:disabled={!page.accessAllowed} href={page.accessAllowed ? page.pageUrl : ""}>
                                <span class="material-symbols-outlined">jump_to_element</span>
                            </a>
                            {#if !page.accessAllowed }
                                <p class="page-no-access">Locked</p>
                            {/if}
                        </div>
                    </AppCard>
                {/each}
            </div>
        {:else}
            <FetchFailed error={$pagesStore} />
        {/if}
    </div>
</PageOutline>

<style>
    .page-container {
        display: flex;
        flex-flow: row wrap;
        gap: 2rem;
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
