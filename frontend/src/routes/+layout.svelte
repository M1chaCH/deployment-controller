<script lang="ts">
    import {isErrorDto} from '$lib/api/api.js';
    import {userStore} from '$lib/api/store';
    import logo from '$lib/assets/michu-tech-icon-black-white.svg';
    import ThemeSwitcher from '$lib/colors/ThemeSwitcher.svelte';
    import ProfileImage from '$lib/ProfileImage.svelte';

    // TODO cleanup
    $:username = $userStore?.mail?.split("@")[0] ?? null

    // TODO add current url to /login?origin ...
</script>

<header>
    <a class="part title-part" href="/">
        <img alt="michu-tech logo small" src={logo} />
        <h1 class="title">michu - <span>TECH</span></h1>
    </a>
    <div class="part">
        <ThemeSwitcher />

        {#if $userStore !== null && !isErrorDto($userStore) }
            <ProfileImage bind:username/>
        {:else }
            <a id="largeLoginButton" class="carbon-button secondary" href="/login">
                Login
                <span class="material-symbols-outlined icon">login</span>
            </a>
            <a id="smallLoginButton" class="icon-button" href="/login">
                <span class="material-symbols-outlined">login</span>
            </a>
        {/if}

    </div>
</header>
<main>
    <slot></slot>
</main>

<style>
    header {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;

        position: sticky;
        top: 0;

        height: 4rem;
        box-sizing: border-box;

        background-color: var(--controller-area-color);
        outline: solid 1px var(--controller-line-color);
    }

    header img {
        height: 100%;
        width: auto;
        aspect-ratio: 1/1;
        box-sizing: border-box;
    }

    header .part {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        height: 100%;
    }

    .title-part {
        padding-right: 1rem;
    }

    .title-part:hover {
        background-color: var(--controller-hover-color);
    }

    .title-part:focus, .title-part:active {
        background-color: var(--controller-focus-color);
    }

    .title {
        margin-left: 1.5rem;
        font-size: 2rem;
        font-weight: 200;
        letter-spacing: -0.05rem;
    }

    .title span {
        font-weight: 300;
    }

    #smallLoginButton {
        display: none;
    }

    @media (max-width: 550px) {
        #smallLoginButton {
            display: inherit;
        }

        #largeLoginButton {
            display: none;
        }
    }

    @media (max-width: 430px) {
        .title {
            font-size: 1.5rem;
            padding-right: 0;
            margin-left: 0.8rem;
        }
    }

    @media (max-width: 350px) {
        .title {
            display: none;
        }
    }
</style>
