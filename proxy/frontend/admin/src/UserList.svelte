<script>
    import { fly } from 'svelte/transition';
    import { getUsers, deleteUser } from "./admin-api.js"
    import UserEditor from './UserEditor.svelte';

    export let apiUrl;
    let userIdToDelete;
    let userIdToEdit;
</script>

{#await getUsers(apiUrl)}
    <p>Loading pages...</p>
{:then users }
    <div class="editable-list">
        {#each users as user, index (index)}
            <div class="row">
                {#if userIdToEdit === user.userId}
                    <UserEditor apiUrl={apiUrl} create={false} user={user} on:cancel={() => userIdToEdit = undefined} />
                {:else}
                    <p class="cell">{user.mail}</p>
                    {#each user.pages as page}
                        <p class="cell" style="display: flex; align-items: center; justify-content: center; gap: 5px;">
                            <span>{page.url}</span>
                            <span class="material-symbols-rounded">{page.hasAccess ? 'check' : 'block'}</span>
                        </p>
                    {/each}

                    <div style="width: 50px; display: flex; align-items: center; justify-content: center;">
                        {#if user.admin}
                            <span class="material-symbols-rounded">shield_person</span>
                        {:else }
                            <span class="material-symbols-rounded">person</span>
                        {/if}
                    </div>

                    <button type="button" class="icon-button" on:click|preventDefault={() => userIdToDelete = userIdToDelete === user.userId ? undefined : user.userId}>
                        <span style="color: var(--michu-tech-foreground);" class="material-symbols-rounded">delete</span>
                    </button>
                    <button type="button" class="icon-button" on:click|preventDefault={() => userIdToEdit = user.userId}>
                        <span style="color: var(--michu-tech-foreground);" class="material-symbols-rounded">edit</span>
                    </button>
                {/if}
            </div>
            {#if user.userId === userIdToDelete}
                <div transition:fly="{{delay: 0, duration: 300, y: -20 }}">
                    <p style="text-align: center;">Do you really want to delete "{user.mail}"?</p>
                    <div style="display: flex; flex-flow: row nowrap; align-items: center; justify-content: space-evenly;">
                        <button type="button" class="button" style="filter: opacity(75%)" on:click|preventDefault={() => deleteUser(apiUrl, user.userId)}>Yes</button>
                        <button type="button" class="button" on:click|preventDefault={() => userIdToDelete = undefined}>Cancel</button>
                    </div>
                </div>
            {/if}
            {#if index < users.length - 1}
                <div class="divider"></div>
            {/if}
        {/each}
    </div>
{:catch error }
    <p>Could not load pages -- {error}</p>
{/await}

<style>

</style>