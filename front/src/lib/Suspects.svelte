<script lang="ts">
    import SuspectSvelte from './Suspect.svelte'
    import type { Suspect } from '$lib/main';

    export let suspects: Suspect[];
    export let gameOver: boolean;
    export let investigationOver: boolean;
    export let answerIsLoading: boolean;
    export let answerFailed: boolean;
</script>

<div class="suspects">
    {#each suspects as suspect}
        <SuspectSvelte suspect={suspect} {gameOver} {investigationOver} {answerIsLoading} {answerFailed} on:suspect_freeing on:suspect_jailing />
    {/each}
</div>

<style>
.suspects {
    justify-content: flex-start;
    display: flex;
    flex-wrap: wrap;
    width: 120vh;
    max-width: 100vw;
    max-height: 80vh;
    flex-shrink: 1;
}
@media screen and (max-width: 600px) {
    .suspects {
        width: unset;
        justify-content: center;
    }
}
</style>