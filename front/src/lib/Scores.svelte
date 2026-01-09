<script lang="ts">
    import { currentGame, hint } from '$lib/stores';
    import { type FinalScore } from '$lib/main';
    import { GetScores, SaveScore } from '$lib/main';
    import { createEventDispatcher, onMount } from 'svelte';
    import { t } from 'svelte-i18n';
	import { goto } from '$app/navigation';

    let name: string;
    let scores: FinalScore[] = [];
    let loading: boolean = true;

    const dispatch = createEventDispatcher();

    // Fetch the scores when the component is mounted
    onMount(async () => {
        try {
            scores = await GetScores();
        } catch (error) {
            console.error('Error fetching scores:', error);
        } finally {
            loading = false;
        }
    });

    function closeScores() {
        dispatch('toggleScores', { scoresVisible: false });
    }

    function gotoNewGame() {
        goto('/new_game');
    }

    // TODO: also set the name to the local storage, here or inside the function
    async function saveScore(e: Event) {
        await SaveScore(name, $currentGame.uuid);
        const target = e.target as HTMLElement | null;
        if (target) {
            target.style.display = "none";
        }
        
        const input = document.getElementById("name_input") as HTMLInputElement | null;
        if (input) {
            const span = document.createElement("span");
            span.textContent = input.value;
            span.id = input.id;
            input.replaceWith(span);
        }

        
    }

    // Helper function to return the position label (medal or rank)
    function getPositionLabel(position: number): string {
        if (position === 1) return '🥇';
        if (position === 2) return '🥈';
        if (position === 3) return '🥉';
        return `${position}.`;
    }

    // Function to check if the current score belongs to the current game
    function isCurrentGame(scoreUUID: string): boolean {
        return scoreUUID === $currentGame.uuid;
    }

    function getHintNewGame() {
        return hint.set("Start a new game and try it again!");
    }
</script>
<div class="infobox_overlay" role="dialog" aria-modal="true">
<div class="infobox">
    <h1>{$t('gameOver.gameOver')}</h1>
    <div class="riptext">{$t('gameOver.riptext')}</div>

    <h2>{$t('gameOver.highScores')}</h2>
    {#if loading}
        {$t('gameOver.loadingScores')}
    {:else}
        <div class="scores">
            {#each scores as score, index}
                {#if index < 10 || isCurrentGame(score.GameUUID)}                
                    {#if index >= 10 && isCurrentGame(score.GameUUID)}
                        <div class="score-item">...</div>
                    {/if}
                    <div class="score-item" class:highlighted={isCurrentGame(score.GameUUID)}>
                        <span class="position">{getPositionLabel(index + 1)}</span>
                        
                        {#if isCurrentGame(score.GameUUID)}
                            <span>
                                <input id="name_input"
                                    bind:value={name}
                                    on:mouseenter={() => hint.set("Inscribe your name to the leaderboards.")}
                                    on:mouseleave={() => hint.set("")}
                                    placeholder="{$t('gameOver.enterName')}"
                                />
                                <button
                                    on:click={saveScore}
                                    on:mouseenter={() => hint.set("Confirm your name and save.")}
                                    on:mouseleave={() => hint.set("")}
                                    >
                                    {$t('buttons.confirm')}
                                </button>
                            </span>
                        {:else}
                            {score.Investigator}
                        {/if}
                        <span class="score">{score.Score}</span>
                    </div>

                {/if}
            {/each}
        </div>
    {/if}
    <button
        on:click={gotoNewGame}
        on:mouseenter={() => getHintNewGame()}
        on:mouseleave={() => hint.set("")}
        >
        {$t('buttons.newGame')}
    </button>
</div>
</div>

<style>
h1 {
    margin: 0;
}

.scores {
    margin: 30px 0;
    display: flex;
    flex-direction: column;
    align-items: center; /* Center align items horizontally */
    text-align: center;  /* Center align text */
}

.score-item {
    max-width: 90%;
    margin-bottom: 8px;
    font-size: 18px;
    display: flex;
    justify-content: space-between;
    align-items: center;  /* Align items vertically in the center */
    width: 100%;  /* Make sure it spans the full width */
}

.score-item input {
    margin: 0 0 0 2rem;
}


.highlighted {
    background-color: rgb(255, 89, 0);
    padding: 5px;
    border-radius: 5px;
}


</style>
