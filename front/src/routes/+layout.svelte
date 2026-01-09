<script lang="ts">
    import { register, init, waitLocale } from 'svelte-i18n';
    import { onMount } from 'svelte';
    import { browser } from '$app/environment';
    import posthog from 'posthog-js'
    import OverlayError from '$lib/OverlayError.svelte';

    export const load = async () => {
    if (browser) {
        console.log("POSTHOG INITED!")
        posthog.init(
        'phc_YvVKqEIOo7uTZMg0AsElZNWqPHguDF0j7eFNEATigLf',
        {
            api_host: 'https://eu.i.posthog.com',
            defaults: '2025-11-30',
            persistence: 'memory',
            autocapture: false,
            person_profiles: 'identified_only', // or 'always' to create profiles for anonymous users as well
        }
        )
    }
    return
    };

    register('en', () => import('$lib/locales/en.json'));
    register('cz', () => import('$lib/locales/cz.json'));
    register('pl', () => import('$lib/locales/pl.json'));

    init({
        fallbackLocale: 'en',
        initialLocale: 'en'
    });

    let isLocaleLoaded = false;

    onMount(async () => {
        await waitLocale(); // Ensure locale is ready before rendering
        isLocaleLoaded = true;
        const player = localStorage.getItem('player');
        console.log("Current player:", player);
        if (browser) {
            console.log("POSTHOG INITED!")
            posthog.init(
                'phc_YvVKqEIOo7uTZMg0AsElZNWqPHguDF0j7eFNEATigLf',
                {
                    api_host: 'https://eu.i.posthog.com',
                    defaults: '2025-11-30',
                    persistence: 'memory',
                    autocapture: false,
                    person_profiles: 'identified_only', // or 'always' to create profiles for anonymous users as well
                }
            )
        }
    });
</script>

{#if isLocaleLoaded}
    <slot />
{:else}
    <p>Loading translations...</p>
{/if}

<OverlayError/>

<style>
:global(:root){
       --bg-color: rgba(27, 38, 54, 1);
}

:global(html) {
    background-color: var(--bg-color);
    text-align: center;
    color: white;
}

:global(body) {
    margin: 0;
    color: white;
    font-family: "Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto",
    "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
    sans-serif;
}

@font-face {
    font-family: "Nunito";
    font-style: normal;
    font-weight: 400;
    src: local(""), url("/nunito-v16-latin-regular.woff2") format("woff2");
}

@font-face {
    font-family: "Bitcount";
    font-style: normal;
    src: local(""), url("/Bitcount-VariableFont_CRSV,ELSH,ELXP,slnt,wght.ttf") format("truetype");
}

:global(h1) {
    margin: 6rem 0 1rem 0;
    font-size: 6rem;
    font-family: "Bitcount";
    font-weight: 390;
    font-variation-settings:
        "slnt" -8,
        "CRSV" 0,
        "ELSH" 0,
        "ELXP" 0;
}

:global(#app) {
    height: 100vh;
    text-align: center;
}

:global(button) {
    display: inline-block;
    outline: 0;
    text-align: center;
    cursor: pointer;
    padding: 5px 10px;
    border: 0;
    color: #fff;
    font-size: 17.5px;
    border: 2px solid transparent;
    border-color: #ffffff;
    color: #ffffff;
    background: transparent;
    transition: background,color .1s ease-in-out;
}
                
:global(button:hover) {
    background-color: #ffffff;
    color: #000000;
}

:global(button:disabled) {
    color: #666666;
    border-color:#666666;
    background-color: unset;
    cursor: not-allowed;
}

:global(button.selected) {
    background: #007bff;
    color: white;
    border-color: #0056b3;
}

:global(header) {
    display: flex;
    justify-content: end;
    padding: 3px 7px 0 0;
}

/* Use display: flow-root to create new Block Formatting Context.
BFC prevents margins to overflow outside main and #app, breaking its 100vh.
*/
:global(main) {
    display: flow-root;
}

:global(footer) {
    margin: 0;
    display: flex;
    justify-content: space-between;
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    box-sizing: border-box;
    padding: 0.2rem 0.7rem 0 7px;
    background-color: rgba(27, 38, 54, 1);
}

:global(a){
    text-decoration: none;
    color: white;
}

:global(.horflip) {
  display: inline-block;
  transform: rotateY(180deg);
}

:global(.menu) {
    display: flex;
    flex-direction: column;
    align-items: center;
}
:globa(.menu > *) {
    max-width: 910px;
}
:global(.menu > button) {
    width: 300px;
    margin: 1rem;
}
:global(.menu > p) {
    margin: 0 0 6rem 0;
    max-width: 900px;
}

:global(.services) {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-top: 5rem;
    gap: 10px;
}
:global(.services > button) {
    min-width: 200px;
}

:global(.infobox_overlay) {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

:global(.infobox) {
    background-color: rgba(27, 38, 54, 1);
    box-sizing: border-box;
    padding: 1.25rem 1.5rem;
    width: min(90%, 800px);
    max-height: 90vh;
    overflow: auto;
    border-radius: 8px;
}


@media screen and (max-width: 600px) {
    /* HOME + NEW GAME */
    :global(h1) {
        font-size: 3rem;
    }
    :global(.menu > p) {
        max-width: 90vw;
    }
    :global(.menu > button) {
        width: 50vw;
        margin: 1rem;
    }    
    :global(.infobox_overlay) {
        align-items: flex-start;
        padding-top: 1rem;
    }
    :global(.infobox) {
        width: calc(100% - 2rem);
        max-height: calc(100vh - 2rem);
        border-radius: 4px;
    }
}

</style>