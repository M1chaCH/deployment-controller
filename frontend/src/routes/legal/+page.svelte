<script lang="ts">
    import {PUBLIC_BACKEND_URL} from '$env/static/public';
    import {isErrorDto} from '$lib/api/open';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import {onMount} from 'svelte';
    import {fly} from 'svelte/transition';

    const maxMessageLength = 1000;
    const mailAddressPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    onMount(() => {
        userStore.subscribe(user => {
            if(!isErrorDto(user) && user?.mail) {
                email = user.mail;
            }
        })
    })

    let email: string;
    let message: string;
    $: contactValid = !sendingLoading && !sentSuccessfully && mailAddressPattern.test(email) && message?.length >= 50 && message?.length <= 1000;
    let sendingLoading = false;
    let sentSuccessfully = false;
    let contactFailed = false;

    async function sendContact() {
        sendingLoading = true;
        sentSuccessfully = false;
        contactFailed = false;

        const response = await fetch(`${PUBLIC_BACKEND_URL}/open/contact`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ mail: email, message: message }),
            credentials: 'include'
        });

        if(response.ok) {
            sentSuccessfully = true;
        } else {
            contactFailed = true;
        }

        setTimeout(() => {
            if(sentSuccessfully) {
                message = "";
            }

            sentSuccessfully = false;
            contactFailed = false;
        }, 5 * 1000);
        sendingLoading = false
    }
</script>

<svelte:head>
    <title>Micha Schweizer @ Legal</title>
</svelte:head>

<PageOutline pageName="Legal">
    <p slot="description">On this page I have documented everything that needs to be done.</p>
    <div slot="content" class="overview-container">
        <div class="content-card">
            <h3>Content</h3>
            <div class="links">
                <a href="#contact">
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                    Contact
                </a>
                <a href="#impressum">
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                    Impressum
                </a>
                <a href="#privacy">
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                    Privacy in my apps
                </a>
                <a href="#app-copyright">
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                    Copyright in my apps
                </a>
            </div>
        </div>
    </div>
</PageOutline>
<div class="legal-content">
    <section id="contact">
        <h3>Contact</h3>
        <p style="margin-bottom: 20px;">Feel free to contact me about any questions or ideas.</p>

        <form on:submit|preventDefault={sendContact}>
            <div class="carbon-input">
                <label for="mail">E-Mail (where I'll respond)</label>
                <input id="mail" class="input" placeholder="E-Mail" bind:value={email}/>
            </div>

            <div class="carbon-input">
                <label for="message">Message</label>
                <textarea cols="12" rows="7" id="message" class="input" placeholder="Message" bind:value={message} maxlength={maxMessageLength}/>
                <span class="count">{ message?.length ?? 0 } / 1000 (min 50)</span>
            </div>

            <div style="display: flex; flex-flow: row-reverse nowrap; justify-content: space-between;">
                <button type="submit" class="carbon-button primary" disabled={!contactValid}>Send</button>

                {#if contactFailed}
                    <MiniNotification message="Failed to send message..." on:close={() => contactFailed = false} />
                    <a href="mailto:admin@michu-tech.com" style="text-decoration: underline; color: var(--michu-tech-accent);">Try this backup.</a>
                {/if}
                {#if sentSuccessfully}
                    <p transition:fly="{{delay: 0, duration: 300, y: -20 }}" class="small-text">↑ Message sent!</p>
                {/if}
            </div>
        </form>
    </section>

    <section id="impressum">
        <h3>Impressum</h3>
        <h4 style="margin: 0;">Micha Schweizer</h4>
        <p>Fliederstrasse 6</p>
        <p>4800 Zofingen</p>
        <p>Schweiz</p>
        <br>
        <p><strong>E-Mail:</strong> <a href="mailto:admin@michu-tech.com">admin@michu-tech.com</a></p>
        <p><strong>Internet:</strong> <a href="#contact" style="text-decoration: underline;">contact</a></p>
    </section>

    <section id="app-privacy">
        <h3>Privacy in my Apps</h3>
        <p>
            At this point, my apps DO NOT collect any data about the user. You can not enter any data into the app and there are no
            background tasks that would collect any data. It's purely readonly, just like a book.
        </p>
    </section>

    <section id="app-copyright">
        <h3>© Copyright in my Apps</h3>
        <h4>Daily Prayer</h4>
        <p>For the daily prayer app we load data from a public bible API.</p>
        <p><a href="https://scripture.api.bible/" target="_blank">API: https://scripture.api.bible/</a></p>
        <br>
        <p>We use this bible translation:</p>
        <p><i>Elberfelder Translation (Version of bibelkommentare.de)</i></p>
        <p>© 2019 by Verbreitung des christlichen Glaubens e.V.</p>
        <br>
        <div class="bible-copyright">
            <p>Dieser Bibeltext ist online verfügbar auf:</p>
            <p><a href="https://www.bibelkommentare.de">www.bibelkommentare.de</a></p>
            <p><strong>Vorwort zur Version von bibelkommentare.de </strong></p>
            <p>Nachdem seit einigen Jahren der Text von 1932 der sogenannten unrevidierten Elberfelder Bibel auf bibelkommentare.de in
               der Bibel mit Suchfunktion und Studienbibel verwendet worden ist, haben wir als Betreiber einige Wortänderungen am Text vorgenommen.</p>
            <p>Zuallererst sei das Wort \"Jehova\" (z.T. auch als \"Jahwe\" in digitalen Übersetzungen bekannt) erwähnt. Das heute in
               Bibelübersetzungen nicht mehr gebräuchliche Wort für JHWH war öfters Anlass zu Kritik und Rückfragen bzgl. der
               Lehrauffassungen von bibelkommentare.de. Die Seite wurde fälschlicherweise mit den Irrlehren einer Sekte in Verbindung
               gebracht. Der Name "Jehova" wurde daher, wie heute in allen Bibelübersetzungen üblich, durch HERR ersetzt. </p>
            <p>Darüber hinaus werden einzelne Wörter, die im Sprachgebrauch nicht mehr üblich sind, durch heute gebräuchliche Synonyme
               ersetzt. Leitfaden bei diesen Änderungen ist oftmals der Duden, Band 1, Die deutsche Rechtschreibung. In der Historie sind die Änderungen ersichtlich. </p>
            <p>Die Verszählung wurde an andere deutsche Bibelübersetzungen angepasst. Die meisten Unterschiede betreffen die Psalmen,
               wo eine vorhandene Überschrift jeweils als erster Vers angegeben wird.</p>
            <p>Wir sind uns der Heiligkeit von Gottes Wort bewusst und schätzen die sorgfältige Arbeit der Brüder, die die Elberfelder
               Bibel vor über einem Jahrhundert übersetzt haben. Die gemachten Änderungen sollen dem heutigen Leser helfen auf weniger veraltete Worte zu stoßen. </p>
            <p>Die Strong-Nummern zum deutschen Bibeltext erfolgten auf Basis der Zuordnung von Alexander vom Stein.</p>
            <p>Das bibelkommentare.de-Team.</p>
        </div>
    </section>
</div>

<style>
    .overview-container {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100%;

        margin: 10vh 0;
    }

    .content-card {
        padding: 2rem;
        box-sizing: border-box;
    }
    .links {
        display: flex;
        flex-flow: column;
        gap: 1rem;
    }

    .links a {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        gap: 1rem;

        padding-top: 1rem;

        transition: transform 120ms ease-out;
    }

    .links a:hover {
        transform: translateX(2rem);
    }

    .legal-content {
        display: flex;
        flex-flow: column;
        gap: 2rem;

        box-sizing: border-box;
        padding: 2rem;
        margin: auto;
        max-width: 1200px;
    }

    #contact form {
        display: flex;
        flex-flow: column;
        gap: 20px;
    }

    #contact form textarea {
        width: 100%;
        resize: vertical;
        min-height: 100px;
    }

    .bible-copyright p {
        padding-bottom: 1rem;
    }
</style>