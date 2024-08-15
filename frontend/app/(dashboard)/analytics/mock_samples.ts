import { agnewsAssociations, agnewsReformulations, agnewsUpvotes } from "./agnews";
import { bookAssociations, goodBookReformulations, badBookReformulations, bookUpvotes } from "./books";

interface Upvote {
    query: string;
    upvote: string;
}

interface Association {
    source: string;
    target: string;
}

interface Reformulation {
    original: string;
    reformulations: string[];
}

export const mockSamples: Record<string, { upvotes: Upvote[], associations: Association[], reformulations: Reformulation[] }> = {
    "default": {
        upvotes: agnewsUpvotes, 
        associations: agnewsAssociations, 
        reformulations: agnewsReformulations,
    },
    "good": {
        upvotes: bookUpvotes, 
        associations: bookAssociations, 
        reformulations: goodBookReformulations,
    },
    "bad": {
        upvotes: bookUpvotes, 
        associations: bookAssociations, 
        reformulations: badBookReformulations,
    },
};

interface Hyperparams {
    numSamples: number;
    maxNewSamples: number;
    probabilityNewSamples: number;
    intervalSeconds: number;
}    

export const rollingSampleParameters: Record<string, {upvotes: Hyperparams, associations: Hyperparams, reformulations: Hyperparams}> = {
    "good": {
        upvotes: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.2,
            intervalSeconds: 2
        },
        associations: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.1,
            intervalSeconds: 3
        },
        reformulations: {
            numSamples: 4,
            maxNewSamples: 2,
            probabilityNewSamples: 0.4,
            intervalSeconds: 2
        },
    },
    "bad": {
        upvotes: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.2,
            intervalSeconds: 2
        },
        associations: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.1,
            intervalSeconds: 3
        },
        reformulations: {
            numSamples: 4,
            maxNewSamples: 2,
            probabilityNewSamples: 0.4,
            intervalSeconds: 2
        },
    },
    "default": {
        upvotes: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.2,
            intervalSeconds: 2
        },
        associations: {
            numSamples: 7,
            maxNewSamples: 3,
            probabilityNewSamples: 0.1,
            intervalSeconds: 3
        },
        reformulations: {
            numSamples: 4,
            maxNewSamples: 2,
            probabilityNewSamples: 0.4,
            intervalSeconds: 2
        },
    }
};