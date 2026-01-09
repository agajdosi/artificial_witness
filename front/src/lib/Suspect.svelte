<script lang="ts">
    import { hint } from '$lib/stores';
    import type { Suspect } from '$lib/main';
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();

    export let suspect: Suspect;
    export let gameOver: boolean;
    export let investigationOver: boolean;
    export let answerIsLoading: boolean;
    export let answerFailed: boolean;

    const imgDir: string = 'suspects/';

    async function selected() {
        if (suspect.Free || suspect.Fled || gameOver || answerIsLoading || answerFailed) return;
        if (investigationOver) { // last suspect = click to jail, new Investigation coming
            dispatch('suspect_jailing', { 'suspect': suspect})
            return
        }
        dispatch('suspect_freeing', { 'suspect': suspect });
    }

    function setHint() {
        if (suspect.Free) return hint.set("Suspect was released.");
        if (answerIsLoading) return hint.set("Before releasing... Wait for the AI to answer the question.");
        if (answerFailed) return hint.set("AI failed to answer - push retry to ask it again.");
        if (suspect.Fled) return hint.set("Criminal was released, game over!");
        if (gameOver) return hint.set("Falsely investigating the innocent.");
        hint.set("Click to release an innocent suspect.");
    }

    $: suspectClasses = [
        "suspect",
        suspect.Free && "free",
        suspect.Fled && "fled",
        answerIsLoading && "waiting",
        answerFailed && "offline",
        investigationOver && !suspect.Free && "to_jail",
        gameOver && !suspect.Fled && !suspect.Free && "accused"
    ].filter(Boolean).join(" ");
</script>

<div
    class={suspectClasses}
    id={suspect.UUID}
    on:click={selected}
    on:keydown={selected}
    on:mouseenter={setHint}
    on:mouseleave={() => hint.set("")}
    aria-disabled={suspect.Free || suspect.Fled || gameOver || answerFailed}
    role="button"
    tabindex="0"
    >
    <div class="suspect-image" style="background-image: url({imgDir+suspect.Image});"></div>
</div>

<style>
    .suspect {
        position: relative;
        height: 21vh;
        width: 21vh;
        margin: 1%;
        cursor: pointer;
        transition: opacity 0.3s ease, filter 0.3s ease;
    }

    .waiting {
        cursor:progress;
    }

    .suspect.free .suspect-image {
        opacity: 0.2;
        filter: grayscale(100%);
        cursor: not-allowed;
    }

    .suspect.fled::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: rgba(255, 0, 0, 0.5); /* Red overlay */
        pointer-events: none; /* Ensure the overlay doesn’t interfere with clicks */
        transition: background-color 0.3s ease; /* Transition for the red overlay */
        opacity: 0; /* Initially hidden */
    }

    .suspect.fled .suspect-image {
        opacity: 0.6;
        filter: grayscale(20%);
        cursor: not-allowed;
    }

    .suspect.fled::before {
        opacity: 1; /* Fade in the red overlay */
    }

    .suspect-image {
        height: 100%;
        width: 100%;
        background-size: cover;
        background-position: center;
        background-repeat: no-repeat;
        transition: opacity 0.3s ease, filter 0.3s ease;
    }

    .suspect.accused{
        cursor: not-allowed;
        filter: invert(65%);
    }

    .suspect.to_jail :hover{
        filter: contrast(200%);
    }

    .offline {
        cursor: not-allowed;
    }

@media screen and (max-width: 600px) {
    .suspect {
        width: 30vw;
        height: 30vw;
        max-height: 10vh;
    }
}
</style>

