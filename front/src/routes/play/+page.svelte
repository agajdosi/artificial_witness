<svelte:head>
    <title>Artificial Witness | Question the AI and Solve the Case</title> 
</svelte:head>

<script lang="ts">
    import { currentGame, hint, selectedModel } from '$lib/stores';
    import { NextRound, EliminateSuspect, GetGame, NextInvestigation, NewGame, type Suspect } from '$lib/main';
    import Suspects from '$lib/Suspects.svelte';
    import History from '$lib/History.svelte';
    import Scores from '$lib/Scores.svelte';
    import Help from '$lib/Help.svelte';
    import OverlayIntro from '$lib/OverlayIntro.svelte';
    import { locale, t } from 'svelte-i18n';
    import MenuTop from '$lib/MenuTop.svelte';
    import { onMount } from 'svelte';
	import Navigation from '$lib/Navigation.svelte';
	import { goto } from '$app/navigation';

    let scoresVisible: boolean = true;
    let helpVisible: boolean = false;
    let overlayConfigVisible: boolean = true;

    onMount(async () => {
        if ($currentGame.uuid == ""){
            const model = $selectedModel ?? 'ollama';
            if (model === '') gotoNewGame()
            try {
                await NewGame(model);
            } finally {
                selectedModel.set(null);
            }
        }
    });

    function gotoNewGame(){
        goto("/new_game");
    }

    function getHintNextQuestion(){
        if ($currentGame.investigation?.rounds?.at(-1)?.answer == "") return hint.set("Wait for the AI to answer the question.")
        if (!$currentGame.investigation?.rounds?.at(-1)?.Eliminations) return hint.set("Eliminate at least 1 suspect before proceeding to next question.");
        return hint.set("Proceed to next question.");
    }

    async function handleSuspectFreeing(event: CustomEvent<{suspect: Suspect}>) {
        console.log("FREEING SUSPECT", event)
        const { suspect } = event.detail;
        try {
            const roundUUID = $currentGame.investigation?.rounds?.at(-1)?.uuid;
            const investigationUUID = $currentGame.investigation?.uuid;
            if (!roundUUID || !investigationUUID) return;
            await EliminateSuspect(suspect.UUID, roundUUID, investigationUUID);
        } catch (error) {
            console.error(`Failed to free suspect ${suspect.UUID}:`, error);
        }
        const game = await GetGame();
        currentGame.set(game);
        console.log(`GAME OVER: ${game.GameOver}`);
    }

    // Scores
    function handleToggleScores(event: CustomEvent<{scoresVisible: boolean}>) {
        scoresVisible = event.detail.scoresVisible;
    }

    //HELP
    function toggleHelp() {
        helpVisible = !helpVisible;
        scoresVisible = false;
    }
    function handleToggleHelp(event: CustomEvent<{helpVisible: boolean}>) {
        helpVisible = event.detail.helpVisible;
    }

    //INTRO
    let introVisible: boolean = false;
    function handleToggleIntro(event: CustomEvent<{introVisible: boolean}>) {
        introVisible = event.detail.introVisible;
    }

    let IdleTimer: NodeJS.Timeout | null = null;
    window.addEventListener('mousemove', resetIdleTimer); // (re)sets IdleTimer
    function resetIdleTimer(): void {
        const msTimeout = 5 * 60 * 1000;
        if (IdleTimer) {
            clearTimeout(IdleTimer);
        }
        IdleTimer = setTimeout(userIsIdle, msTimeout);
    }
    function userIsIdle() {
        introVisible = true;
    }

</script>

<div class="top">
    <div class="top-left">
        <div class="main">
        {#if $currentGame.investigation?.InvestigationOver}
            <div class="jailtime">
                {$t('arrest')}
            </div>
        {:else}
            <div
                class="question"
                role="tooltip"
                on:mouseenter={() => hint.set("A question about the wanted person, answered by an AI witness.")}
                on:mouseleave={() => hint.set("")}
                >
                {$currentGame.investigation?.rounds?.length}.
                {#if $locale == "cz"}
                    {$currentGame.investigation?.rounds?.at(-1)?.Question?.Czech}
                {:else if $locale == "pl"}
                    {$currentGame.investigation?.rounds?.at(-1)?.Question?.Polish}
                {:else}
                    {$currentGame.investigation?.rounds?.at(-1)?.Question?.English}
                {/if}
            </div>
            {#if $currentGame.investigation?.rounds?.at(-1)?.answer == ""}
                <div class="waiting"
                    role="tooltip"
                    on:mouseenter={() => hint.set("Waiting for the AI witness to answer the question.")}
                    on:mouseleave={() => hint.set("")}
                    >
                    *{$t('thinking')}*
                </div>
            {:else}
                <div class="answer"
                    role="tooltip"
                    on:mouseenter={() => hint.set("The AI witness' response to the question about the wanted person.")}
                    on:mouseleave={() => hint.set("")}
                    >
                    {$t($currentGame.investigation?.rounds?.at(-1)?.answer?.toLowerCase() || '') || $currentGame.investigation?.rounds?.at(-1)?.answer?.toLowerCase() || ''}!
                </div>
            {/if}
        {/if}
        </div>
        <div class="instruction">
            {#if $currentGame.investigation?.InvestigationOver}
                {$t('arrestInstruction')}
            {:else if $currentGame.investigation?.rounds?.at(-1)?.answer != ""}
                {#if $currentGame.investigation?.rounds?.at(-1)?.answer?.toLowerCase() == "yes"}{$t('release-no')}
                {:else}{$t('release-yes')}
                {/if}
            {:else}
                {$t('waiting')}...
            {/if}
        </div>
    </div>
    <div class="top-right"
        role="tooltip"
        on:mouseenter={() => hint.set("Switch language of the user interface.")}
        on:mouseleave={() => hint.set("")}
        >
        <MenuTop bind:overlayConfigVisible={overlayConfigVisible}/>
    </div>
</div>

<div class="middle">
    <div class="left">
        <Suspects
            suspects={$currentGame.investigation?.suspects || []}
            gameOver={$currentGame.GameOver}
            investigationOver={$currentGame.investigation?.InvestigationOver}
            answerIsLoading={$currentGame.investigation?.rounds?.at(-1)?.answer == ""}
            on:suspect_freeing={handleSuspectFreeing}
            on:suspect_jailing={NextInvestigation}
        />

        <div class="actions">
            {#if !$currentGame.investigation?.InvestigationOver}
                {#if $currentGame.GameOver}
                    <button
                        on:click={gotoNewGame}
                        on:mouseenter={() => hint.set("Start a new game and try it again!")}
                        on:mouseleave={() => hint.set("")}
                        >
                        {$t('buttons.newGame')}
                    </button>
                {:else}
                <button
                    class="next-round"
                    on:click={NextRound}
                    on:mouseenter={() => getHintNextQuestion()}
                    on:mouseleave={() => hint.set("")}
                    disabled={!$currentGame.investigation?.rounds?.at(-1)?.Eliminations || $currentGame.GameOver }
                    aria-disabled="{!$currentGame.investigation?.rounds?.at(-1)?.Eliminations || $currentGame.GameOver ? 'true': 'false'}"
                    >
                    {$t('buttons.nextQuestion')}
                </button>
                {/if}
            {/if}
        </div>
    </div>

    <div class="right">
        <History/>
    </div>
</div>

<div class="bottom">
    <div class="help">
        <Navigation/>
        <button on:click={toggleHelp} class="langbtn">{$t('buttons.help')}</button>
    </div>
    <div class="hint">
        <span>{$hint}</span>
    </div>
    <div class="stats">
        <div
            role="tooltip"
            on:mouseenter={() => hint.set("Successfully finish the investigation to get into higher level.")}
            on:mouseleave={() => hint.set("")}
            >
            level: {$currentGame.level}
        </div>
        <div
            role="tooltip"
            on:mouseenter={() => hint.set("Your current score. Free innocent suspects and finish the investigation to get more points.")}
            on:mouseleave={() => hint.set("")}
            >
            score: {$currentGame.Score}
        </div>
    </div>
</div>

{#if $currentGame.GameOver && scoresVisible}
    <Scores on:toggleScores={handleToggleScores}/>    
{/if}

{#if helpVisible}
    <Help on:toggleHelp={handleToggleHelp}/>    
{/if}

{#if introVisible}
    <OverlayIntro on:toggleIntro={handleToggleIntro}/>
{/if}

<style>
.middle {
    display: flex;
}
.left .actions {
    padding: 2rem 0;
}
.right {
    padding: 0.2rem 0 0 0;
    width: 100%;
    max-height: 73vh;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
}

.bottom {
    display: flex;
    justify-content: space-between;
    position: absolute;
    bottom: 0;
    width: calc(100vw - 2.5rem);
    padding: 0 0.5rem 0 7px;
}
.stats {
    display: flex;
    gap: 1rem;
}

.top {
    width: 100vw;
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.top .main {
    padding: 1rem 0 0 0;
    display: flex;
    flex-direction: row;
    gap: 1rem;
    margin: 0 0 0 1rem;
    font-size: 2rem;
}

.top .instruction {
    font-size: 1.2rem;
    display: flex;
    margin: 0 0 0 1.1rem;
}

.top-right {
    padding: 3px 7px 0 0;
}

.answer {
    text-transform: uppercase;
}

.langbtn {
    all: unset;
    text-decoration: underline;
    min-width: 20px;
}
.langbtn:hover{
    cursor: pointer;
}

button:disabled{
    cursor: wait;
}

button.next-round:disabled {
    cursor: not-allowed;
}

.help {
    display: flex;
}

@media screen and (max-width: 600px) {
    .top .main{
        flex-direction: column;
        min-height: 10vh;
        padding: 4px 4px;
        margin: 0;
        gap: 0.2rem;
        font-size: 1.1rem;
    }
    .top .instruction {
        font-size: 0.9rem;
        color: #666666;
        justify-content: center;
    }
    .top .answer {
        align-self:center;
    }
    .top-right {
        display: none;
    }
    .middle {
        flex-direction: column;
    }
    .hint {
        display: none;
    }
    .right {
        flex-direction: column-reverse;
        padding: 0 1rem;
    }
}

</style>
