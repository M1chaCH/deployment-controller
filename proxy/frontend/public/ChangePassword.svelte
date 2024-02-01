<script>
    import { fly } from 'svelte/transition';
    import {createEventDispatcher} from 'svelte';

    export let saveText = "Save";
    export let showCancel = true;
    export let showTitle = true;
    export let errorText = "";

    export let mail;
    let oldPassword;
    let password;
    $: invalid = !mail || !oldPassword || !password;

    const dispatch = createEventDispatcher();

    function dispatchCancel() {
        errorText = "";
        dispatch("cancel", {});
    }

    function dispatchSave() {
        errorText = "";
        dispatch("save", {mail, oldPassword, password});
    }
</script>

<form class="change-password-form">
    {#if showTitle}
        <h2>Change password</h2>
    {/if}
    <input type="text" class="input" bind:value={mail}/>
    <input type="password" class="input" placeholder="Old password" bind:value={oldPassword} autocomplete="current-password"/>
    <input type="password" class="input" placeholder="New password" bind:value={password} autocomplete="new-password"/>

    {#if !!errorText}
        <p transition:fly="{{delay: 0, duration: 300, y: -20 }}">{errorText}</p>
    {/if}

    <button type="submit" class="button" on:click|preventDefault={dispatchSave} disabled={invalid}>{saveText}</button>
    {#if showCancel}
        <button class="button" on:click|preventDefault={dispatchCancel}>Cancel</button>
    {/if}
</form>

<style>
    .change-password-form {
        display: flex;
        flex-flow: column;
        gap: 10px;

        align-items: center;
        justify-content: center;

        width: 100%;
        max-width: 84vw;
        height: 100%;
    }

    input, button {
        min-width: 220px;
        width: 80vw;
        max-width: 360px;
    }
</style>