<script lang="ts">
    import {registerCloseBackdrop} from '$lib';
    import type {UserInfoDto} from '$lib/api/open';
    import {createEventDispatcher, onMount} from 'svelte';

    export let user: UserInfoDto;

    const dispatch = createEventDispatcher()
    const close = () => dispatch("close");
    const changePassword = () => dispatch("changePassword")

    onMount(() => {
        registerCloseBackdrop(close, false);
    })
</script>

<div class="backdrop">
    <div class="popover">
        <div class="labeled-value">
            <label for="username">E-Mail</label>
            <p id="username">{user.mail}</p>
        </div>
        {#if user.onboard}
            <button class="wrapper-button labeled-value" on:click={changePassword} style="width: 100%;">
                <p>Change password</p>
            </button>
        {/if}
        {#if !user.onboard}
            <a class="labeled-value" href="/onboarding">
                <p>Onboard</p>
            </a>
        {/if}
        {#if user.admin}
            <a class="labeled-value" href="/admin">
                <p>Administration</p>
            </a>
        {/if}
    </div>
</div>

<style>
.popover {
    position: fixed;
    top: 4rem;
    right: 0;
    min-width: 220px;

    background-color: var(--controller-area-color);
}
</style>