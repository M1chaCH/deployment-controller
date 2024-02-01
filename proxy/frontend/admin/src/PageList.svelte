<script>
    import { fly } from 'svelte/transition';
    import { getPages, deletePage } from "./admin-api.js"
    import PageEditor from './PageEditor.svelte';

    export let apiUrl;
    let pageIdToDelete;
    let pageIdToEdit;
</script>

{#await getPages(apiUrl)}
    <p>Loading pages...</p>
{:then pages }
    <div class="editable-list">
        {#each pages as page, index (index)}
            <div class="row">
                {#if pageIdToEdit === page.id}
                    <PageEditor apiUrl={apiUrl} create={false} page={page} on:cancel={() => pageIdToEdit = undefined} />
                {:else}
                    <p class="cell">{page.id}</p>
                    <p class="cell">{page.title}</p>
                    <p class="cell">{page.url}</p>
                    <p class="cell" style="flex: 2;">{page.description}</p>

                    <div style="width: 50px; display: flex; align-items: center; justify-content: center;">
                        {#if page.privatePage}
                            <span class="material-symbols-rounded">encrypted</span>
                        {:else }
                            <span class="material-symbols-rounded">lock_open</span>
                        {/if}
                    </div>

                    <button type="button" class="icon-button" on:click|preventDefault={() => pageIdToDelete = pageIdToDelete === page.id ? undefined : page.id}>
                        <span style="color: var(--michu-tech-foreground);" class="material-symbols-rounded">delete</span>
                    </button>
                    <button type="button" class="icon-button" on:click|preventDefault={() => pageIdToEdit = page.id}>
                        <span style="color: var(--michu-tech-foreground);" class="material-symbols-rounded">edit</span>
                    </button>
                {/if}
            </div>
            {#if page.id === pageIdToDelete}
                <div transition:fly="{{delay: 0, duration: 300, y: -20 }}">
                    <p style="text-align: center;">Do you really want to delete "{page.url}"?</p>
                    <div style="display: flex; flex-flow: row nowrap; align-items: center; justify-content: space-evenly;">
                        <button type="button" class="button" style="filter: opacity(75%)" on:click|preventDefault={() => deletePage(apiUrl, page.id)}>Yes</button>
                        <button type="button" class="button" on:click|preventDefault={() => pageIdToDelete = undefined}>Cancel</button>
                    </div>
                </div>
            {/if}
            {#if index < pages.length - 1}
                <div class="divider"></div>
            {/if}
        {/each}
    </div>
{:catch error }
    <p>Could not load pages -- {error}</p>
{/await}

<style>

</style>