<script>
    import { addPage, editPage } from "./admin-api.js"
    import {createEventDispatcher} from 'svelte';
    // import { v4 as uuid } from 'uuid';

    export let apiUrl;
    export let page = undefined;
    export let create = true;
    $: save = create ? addPage : editPage;

    const dispatch = createEventDispatcher();

    let pageId = page?.id;
    let pageTitel = page?.title;
    let pageDesc = page?.description;
    let pageUrl = page?.url;
    let pagePrivate = page ? page.privateAccess : true;
    $: invalidPage = !pageTitel || !pageDesc || !pageUrl || !pageUrl.startsWith("/");
    $: pageId = pageId?.toLowerCase();

    function dispatchCancel() {
        dispatch("cancel", {});
    }
</script>

{#if create}
    <h3>Add Page</h3>
{/if}
<form>
    <div class="labeled-input">
        <label for="page-id">Id (lower)</label>
        <input id="page-id" class="input" type="text" placeholder="app" bind:value={pageId}/>
    </div>
    <div class="labeled-input">
        <label for="page-title">Title</label>
        <input id="page-title" class="input" type="text" placeholder="My App" bind:value={pageTitel}/>
    </div>

    <div class="labeled-input">
        <label for="page-desc">Description</label>
        <input id="page-desc" class="input" type="text" placeholder="This is my app (:" bind:value={pageDesc}/>
    </div>

    <div class="labeled-input">
        <label for="page-url">Url</label>
        <input id="page-url" class="input" type="text" placeholder="/url" bind:value={pageUrl}/>
    </div>

    <div class="labeled-input">
        <label for="page-private">Private</label>
        <input id="page-private" type="checkbox" class="checkbox" bind:checked={pagePrivate}/>
    </div>

    <button type="submit" class="button" on:click|preventDefault={() => save(apiUrl, pageId, pageUrl, pageTitel, pageDesc, pagePrivate)} disabled={invalidPage}>
        {create ? 'Add' : 'Save'}
    </button>
    {#if !create}
        <button class="button" on:click|preventDefault={() => dispatchCancel()}>
            Cancel
        </button>
    {/if}
</form>

<style>
    form {
        display: flex;
        flex-flow: row wrap;
        gap: 10px;
    }

    .labeled-input {
        flex: 1;
        min-width: 250px;
        max-width: 500px;
    }

    form button {
        margin-left: auto;
        align-self: center;
    }

    #page-id {
        text-transform: lowercase;
    }
</style>