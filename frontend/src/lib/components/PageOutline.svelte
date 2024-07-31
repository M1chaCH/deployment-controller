<script lang="ts">
    import logo from '$lib/assets/michu-tech-icon.svg';

    export let pageName: string;
</script>

<section class="landing-page">
    <div class="landing-page-side welcome">
        <div class="logo">
            <img src={logo} alt="michu-tech logo">
            <h2>MICHU - TECH</h2>
        </div>
        <h1>Micha Schweizer <span style="color: var(--michu-tech-accent)">@</span>&nbsp;{pageName}</h1>
        <slot name="description"></slot>
    </div>
    <div class="landing-page-side pages">
        <slot name="content"></slot>
    </div>

</section>

<style>
    .landing-page {
        /* screen - header size */
        min-height: calc(100vh - 4rem);

        display: flex;
        flex-flow: row wrap;
        gap: 0;
    }

    @media (max-width: 1200px) {
        .landing-page {
            flex-direction: column;
        }
    }

    .landing-page-side.welcome {
        flex: 1 1 calc(40% - 2.5rem);
        min-width: 250px;
        display: flex;
        flex-flow: column;
        justify-content: flex-start;

        background-color: color-mix(in srgb, var(--controller-area-color) 60%, var(--michu-tech-background));
        border-right: 1px solid var(--controller-line-color);
        padding: 20vh 0.5rem 10vh 2rem;

        box-sizing: border-box;
    }

    .logo {
        display: flex;
        flex-flow: row nowrap;
        gap: 1rem;

        align-items: center;

        margin-bottom: 2rem;
    }

    .logo img {
        height: 72px;
        width: auto;
        aspect-ratio: 1/1;
    }

    .landing-page-side h1 {
        margin-bottom: 0.5rem;
    }

    .landing-page-side.pages {
        flex: 1 1 calc(60% - 4rem);
        min-width: 250px;
        padding: 2rem;

        /*container: pages / size;*/
    }

    .welcome {
        animation: welcome-fly-in;
        animation-duration: 400ms;
        animation-timing-function: ease-out;
    }

    @keyframes welcome-fly-in {
        0% {
            transform: translateX(-60%);
        }
        100% {
            transform: translateX(0);
        }
    }

    .pages {
        transform: translateX(-150%);
        z-index: -1;
        animation: pages-fly-in;
        animation-duration: 800ms;
        animation-delay: 120ms;
        animation-timing-function: ease-out;
        animation-fill-mode: forwards;
    }

    @keyframes pages-fly-in {
        0% {
            z-index: -1;
            transform: translateX(-150%);
        }
        99% {
            z-index: -1;
        }
        100% {
            z-index: unset;
            transform: translateX(0);
        }
    }
</style>