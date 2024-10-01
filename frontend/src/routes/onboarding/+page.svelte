<script lang="ts">

    import {goto} from '$app/navigation';
    import {PUBLIC_BACKEND_URL} from '$env/static/public';
    import {isErrorDto, putChangePassword} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import TokenInput from '$lib/components/TokenInput.svelte';
    import {onMount} from 'svelte';

    let mail: string = "";
    let oldPassword: string;
    let password: string;
    let token = "";
    $: invalid = !mail || !oldPassword || !password || oldPassword === password || !token || token.length < 6;
    let onboardingFailed = false;

    onMount(() => {
        userStore.subscribe(user => {
            if(!isErrorDto(user)) {
                mail = user?.mail ?? mail

                if(user?.onboard) {
                    goto("/");
                }
            }
        })
    })

    async function onboard() {
        onboardingFailed = false;
        if(!invalid && !isErrorDto($userStore)) {
            const result = await putChangePassword({
                                                       userId: $userStore!.userId,
                                                       newPassword: password,
                                                       oldPassword,
                                                       token,
                                                   }, true)

            if(isErrorDto(result)) {
                onboardingFailed = true;
            } else {
                // want to do a page reload, to update caches
                location.href = "/";
                return
            }
        }
    }
</script>

<svelte:head>
    <title>Micha Schweizer @ Onboarding</title>
</svelte:head>

<PageOutline pageName="Onboarding">
    <div slot="description">
        <p>Please change your password and setup your two factor login to activate your account.</p>
        <p class="subtext">Your password must be at least 8 characters long and must match the following validations. [ >= 8 Letters, min. 1 number, min. 1 a-z, min. 1 A-Z ]</p>
    </div>
    <div slot="content" class="page" id="onboarding">
        <div class="content-card">
            <form class="onboarding-form">
                <div class="onboarding-form-side">
                    <h4>Change password</h4>
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
                </div>
                <div class="onboarding-form-side">
                    <h4>Create Token</h4>
                    <img src={PUBLIC_BACKEND_URL + "/open/login/onboard/img"} alt="onboarding token"/>
                    <p>Please scan this QR-Code with a two factor authenticator app. Every time you login with a new device you will have to use this code to login.</p>
                    <TokenInput on:input={(e) => token = e.detail.value}/>
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

    .content-card {
        max-width: 100%;
        container: onboarding / inline-size;
    }

    .onboarding-form {
        display: grid;
        grid-template-columns: repeat(2, calc(50% - 0.5rem));
        gap: 1rem;

        padding: 1rem 2rem;
        box-sizing: border-box;
    }

    .onboarding-form-side {
        min-width: 220px;
        width: 100%;
    }

    .onboarding-form-side img {
        width: 100%;
        height: 200px;
        object-fit: contain;
        background-color: white;
        border-left: 2px solid var(--michu-tech-accent);
    }

    @container onboarding (max-width: 600px) {
        .onboarding-form {
            grid-template-columns: 100% !important;
        }
    }
</style>
