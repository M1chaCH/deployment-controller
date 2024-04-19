<script>
    import { fly } from 'svelte/transition';

    const maxMessageLength = 1000;
    const mailAddressPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    export let apiUrl;
    let email;
    let message;
    $: contactValid = mailAddressPattern.test(email) && message?.length >= 50 && message?.length <= 1000;
    let sentSuccessfully = false;
    let errorMessage;

    async function sendContact() {
        sentSuccessfully = false;
        errorMessage = undefined;

        const response = await fetch(`${apiUrl}/contact`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ mail: email, message: message }),
            credentials: 'include' // Include cookies in subsequent requests
        });

        if(response.ok) {
            sentSuccessfully = true;
        } else {
            errorMessage = `Failed to send message: ${response.status} - ${response.statusText}`
        }

        setTimeout(() => {
            if(sentSuccessfully) {
                message = "";
            }

            sentSuccessfully = false;
            errorMessage = "";
        }, 5 * 1000);
    }
</script>

<svelte:head>
    <title>Micha Schweizer @ Legal</title>
</svelte:head>

<main>
    <h1>Micha Schweizer @ Legal</h1>
    <p class="page-description">On this page I have documented everything that needs to be done.</p>

    <h2>Content</h2>
    <ul>
        <li><a href="#impressum">Impressum</a></li>
        <li><a href="#contact">Contact</a></li>
        <li><a href="#app-privacy">Privacy in my Apps</a></li>
        <li><a href="#app-copyright">Copyright in my Apps</a></li>
    </ul>

    <div class="legal-content">
        <section id="impressum">
            <h2>Impressum</h2>
            <h4 style="margin: 0;">Micha Schweizer</h4>
            <p>Fliederstrasse 6</p>
            <p>4800 Zofingen</p>
            <p>Schweiz</p>
            <br>
            <p><strong>E-Mail:</strong> <a href="mailto:admin@michu-tech.com">admin@michu-tech.com</a></p>
            <p><strong>Internet:</strong> <a href="#contact">Contact</a></p>
        </section>

        <section id="contact">
            <h2>Contact</h2>
            <p style="margin-bottom: 20px;">Feel free to contact me about any questions or ideas.</p>

            <form on:submit|preventDefault={sendContact}>
                <div class="labeled-input">
                    <label for="mail">E-Mail (where I'll respond)</label>
                    <input id="mail" class="input" placeholder="E-Mail" bind:value={email}/>
                </div>

                <div class="labeled-input">
                    <label for="message">Message</label>
                    <textarea cols="12" rows="7" id="message" class="input" placeholder="Message" bind:value={message} maxlength={maxMessageLength}/>
                    <span class="count">{ message?.length ?? 0 } / 1000 (min 50)</span>
                </div>

                <div style="display: flex; flex-flow: row-reverse nowrap; justify-content: space-between;">
                    <button type="submit" class="button" disabled={!contactValid}>Send</button>

                    {#if errorMessage}
                        <p transition:fly="{{delay: 0, duration: 300, y: -20 }}" class="small-text error">{errorMessage}</p>
                    {/if}
                    {#if sentSuccessfully}
                        <p transition:fly="{{delay: 0, duration: 300, y: -20 }}" class="small-text">↑ Message sent!</p>
                    {/if}
                </div>
            </form>
        </section>

        <section id="app-privacy">
            <h2>Privacy in my Apps</h2>
            <p>
                At this point, my apps DO NOT collect any data about the user. You can not enter any data into the app and there are no
                background tasks that would collect any data. It's purely readonly, just like a book.
            </p>
        </section>

        <section id="app-copyright">
            <h2>© Copyright in my Apps</h2>
            <h3>Daily Prayer</h3>
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
</main>

<style>
    .legal-content {
        width: clamp(100px, 100%, 1200px);

        display: flex;
        flex-flow: column;
        gap: 28px;

        box-sizing: border-box;
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

    .labeled-input .count {
        position: absolute;
        right: 25px;
        bottom: 5px;

        font-weight: 300;
        font-size: 12px;
    }

    p {
        margin: 0;
    }

    .small-text.error {
        color: var(--michu-tech-warn);
        font-weight: 500;
        font-style: unset;
    }

    .bible-copyright p {
        padding-bottom: 10px;
    }
</style>
