<script lang="ts">
    import {registerCloseBackdrop} from '$lib';
    import {isErrorDto, putChangePassword} from '$lib/api/open';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import {createEventDispatcher, onMount} from 'svelte';

    export let userId: string;
    export let targetOtherUserEmail: string | undefined = undefined;
    let oldPassword: string;
    let password: string;
    $: invalid = (!targetOtherUserEmail && !oldPassword) || !password || oldPassword === password;

    let changeFailed = false;

    const dispatch = createEventDispatcher();
    const close = () => dispatch("close")
    onMount(() => {
        registerCloseBackdrop(close, true);
    })

    async function changePassword() {
        if(!invalid) {
            changeFailed = false
            const result = await putChangePassword({userId, oldPassword, newPassword: password})

            if(isErrorDto(result)) {
                changeFailed = true
            } else {
                close();
            }
        }
    }
</script>

<div class="backdrop">
    <div class="content-card">
        <form class="change-password-form">
            <h4>Change Password</h4>
            {#if targetOtherUserEmail}
                <p class="subtext">For {targetOtherUserEmail}</p>
            {/if}
            <p>The new password must be different to the first one and must match the following criteria. [ >= 8 Letters, min. 1 number, min. 1 a-z, min. 1 A-Z ]</p>
            {#if !targetOtherUserEmail}
                <div class="carbon-input">
                    <label for="oldPassword">Old password</label>
                    <input autocomplete="current-password" id="oldPassword" bind:value={oldPassword} type="password"/>
                </div>
            {/if}
            <div class="carbon-input">
                <label for="newPassword">New password</label>
                <input id="newPassword" bind:value={password} type="password" autocomplete="new-password"/>
            </div>

            {#if changeFailed}
                <MiniNotification message="Failed to change password." on:close={() => changeFailed = false} />
            {/if}
        </form>
        <div class="controls">
            <button class="carbon-button secondary" on:click|preventDefault={close}>
                <span class="material-symbols-outlined icon">arrow_left_alt</span>
                Cancel
            </button>
            <button class="carbon-button primary" on:click|preventDefault={changePassword} disabled={invalid}>
                Save
                <span class="material-symbols-outlined icon">arrow_right_alt</span>
            </button>
        </div>
    </div>
</div>

<style>
    .change-password-form {
        padding: 1rem 2rem;
    }

    .change-password-form p {
        margin-bottom: 1rem;
    }
</style>