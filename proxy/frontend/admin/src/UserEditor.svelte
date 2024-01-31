<script>
    import { addUser, editUser } from "./admin-api.js"
    import {createEventDispatcher} from 'svelte';
    import { v4 as uuid } from 'uuid';

    export let apiUrl;
    export let user = undefined;
    export let create = true;

    const dispatch = createEventDispatcher();

    let userId = create ? uuid() : user?.userId;
    let mail = user?.mail;
    let password;
    let admin = user ? user.admin : false;
    let pageAccessToAdd = [];
    let pageAccessToRemove = [];
    $: userInvalid = !userId || !mail || create && !password;

    const pagesToEdit = (user?.pages ? JSON.parse(JSON.stringify(user.pages)) : []).filter(p => p.privatePage);

    function dispatchCancel() {
        dispatch("cancel", {});
    }

    function save() {
        if(create) {
            addUser(apiUrl, userId, mail, password, admin, []);
        } else {
            editUser(apiUrl, userId, password, admin, pageAccessToAdd, pageAccessToRemove);
        }
    }

    function togglePageOfUser(page) {
        if(page.hasAccess) {
            if(!user.pages.find(p => p.pageId === page.pageId).hasAccess) {
                pageAccessToAdd.push(page.pageId);
            }
            pageAccessToRemove.splice(pageAccessToRemove.indexOf(p => p.pageId === page.pageId), 1);
        } else if(!page.hasAccess) {
            if(user.pages.find(p => p.pageId === page.pageId).hasAccess) {
                pageAccessToRemove.push(page.pageId);
            }
            pageAccessToAdd.splice(pageAccessToAdd.indexOf(p => p.pageId === page.pageId), 1);
        }
    }
</script>

{#if create}
    <h3>Add User</h3>
{/if}
<form>
    <div class="labeled-input">
        <label for="user-mail">E-Mail</label>
        <input id="user-mail" autocomplete="off" class="input" type="email" placeholder="mail@test.com" bind:value={mail} disabled={!create}/>
    </div>
    <div class="labeled-input">
        <label for="user-password">Password</label>
        <input id="user-password" class="input" type="password" placeholder="12345678" bind:value={password}/>
    </div>

    <div class="labeled-input">
        <label for="user-admin">Admin</label>
        <input id="user-admin" type="checkbox" class="checkbox" bind:checked={admin}/>
    </div>

    {#if !create}
        {#each pagesToEdit as page}
            <div class="labeled-input" style="min-width: 100px;">
                <label>{page.url}</label>
                <input type="checkbox" class="checkbox" bind:checked={page.hasAccess} on:change={() => togglePageOfUser(page)}/>
            </div>
        {/each}
    {/if}

    <button type="submit" class="button" on:click|preventDefault={() => save()} disabled={userInvalid}>
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
</style>