<script lang="ts">

    import {isErrorDto, putChangePassword} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import {onMount} from 'svelte';

    let mail: string = "";
    let oldPassword: string;
    let password: string;
    $: invalid = !mail || !oldPassword || !password || oldPassword === password;
    let onboardingFailed = false;

    onMount(() => {
        userStore.subscribe(user => {
            mail = !isErrorDto(user) ? user?.mail ?? mail : mail
        })
    })

    async function onboard() {
        onboardingFailed = false;
        if(!invalid && !isErrorDto($userStore)) {
            const result = await putChangePassword({
                                                       userId: $userStore!.userId,
                                                       newPassword: password,
                                                       oldPassword,
                                                   }, true)

            if(isErrorDto(result)) {
                onboardingFailed = true;
            } else {
                location.href = "/"
                return
            }
        }
    }
</script>

<PageOutline pageName="Onboarding">
    <div slot="description">
        <p>Please change your password to activate your account.</p>
        <p class="subtext">Your password must be at least 8 characters long and must match the following validations. [ >= 8 Letters, min. 1 number, min. 1 a-z, min. 1 A-Z ]</p>
    </div>
    <div slot="content" class="page" id="onboarding">
        <div class="content-card">
            <form>
                <div class="carbon-input">
                    <label for="mail">E-Mail</label>
                    <input id="mail" type="email" bind:value={mail}/>
                </div>
                <div class="carbon-input">
                    <label for="oldPassword">Old Password</label>
                    <input id="oldPassword" type="password" bind:value={oldPassword} autocomplete="current-password"/>
                </div>
                <div class="carbon-input">
                    <label for="password">Password</label>
                    <input id="password" type="password" bind:value={password} autocomplete="new-password"/>
                </div>
            </form>
            {#if onboardingFailed}
                <MiniNotification message="Something went wrong, does your password match the guidelines?" on:close={() => onboardingFailed = false} />
            {/if}
            <div class="controls">
                <a class="carbon-button secondary" href="/">
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Back
                </a>
                <button class="carbon-button primary" on:click|preventDefault={onboard} disabled={invalid}>
                    Save
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
</PageOutline>

<style>
    .page {
        display: flex;
        flex-flow: column;
        /* page - header height */
        min-height: calc(100vh - 4rem);
        align-items: center;
        justify-content: center;
    }

    .content-card form {
        padding: 1rem 2rem;
    }
</style>
