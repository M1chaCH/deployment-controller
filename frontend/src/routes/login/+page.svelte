<script lang="ts">
    import {goto} from '$app/navigation';
    import {type LoginState, postLogin} from '$lib/api/auth.js';
    import {isErrorDto} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import TokenInput from '$lib/components/TokenInput.svelte';
    import {onMount} from 'svelte';

    onMount(() => {
        userStore.subscribe(user => {
            if(user && !isErrorDto(user)) {
                mail = user.mail;

                if(user.loginState === "logged-in") {
                    goto(parseTargetUrl())
                    return
                }

                if(user.loginState === "onboarding-waiting") {
                    goto("/onboarding");
                    return
                }

                mfa = user.loginState === "two-factor-waiting";
            }
        })
    })

    let mfa = false;
    let failed = false;
    $: formValid = mail?.length > 3 && password && (!mfa || (token?.length === 6 && !isNaN(parseInt(token))));

    let mail = "";
    let password = "";
    let token = "";

    async function login() {
        failed = false;
        if(!formValid) {
            failed = true;
            return
        }

        let state: LoginState;
        try {
             state = await postLogin(mail, password, token);
        } catch (e) {
            console.error(e);
            state = "logged-out"
        }

        switch (state) {
            case 'logged-in':
                await goto(parseTargetUrl())
                return
            case 'onboarding-waiting':
                await goto("/onboarding")
                return
            case 'logged-out':
                failed = true
                break
            case 'two-factor-waiting':
                mfa = true
                break
        }
    }

    function parseTargetUrl() {
        const props = new URLSearchParams(window.location.search);
        const origin = props.get("origin");

        if(origin) {
            return `https://${origin}`;
        }

        return "/";
    }
</script>

<PageOutline pageName="Login">
    <div slot="description">
        <p>Logging in can grant you access to more services.</p>
        <p class="subtext">An account needs to be created by the admin.</p>
    </div>
    <div slot="content" class="page" id="login">
        <div class="content-card">
            <form class="login-inputs">
                <div class="carbon-input">
                    <label for="mail">E-Mail</label>
                    <input id="mail" type="email" bind:value={mail}/>
                </div>
                <div class="carbon-input">
                    <label for="password">Password</label>
                    <input id="password" type="password" bind:value={password}/>
                </div>
                {#if mfa}
                    <TokenInput on:input={e => token = e.detail.value}/>
                    <p class="subtext">New Device, TwoFactor authentication required.</p>
                {/if}
            </form>
            {#if failed}
                <MiniNotification message="Login failed!" on:close={() => failed = false} />
            {/if}
            <div class="controls">
                <a class="carbon-button flat" href="mailto:admin@michu-tech.com?subject=Request Account">
                    Request Account
                </a>
                <button class="carbon-button primary" on:click|preventDefault={login} disabled={!formValid}>
                    Login
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

    .login-inputs {
        padding: 1rem 2rem;
    }
</style>