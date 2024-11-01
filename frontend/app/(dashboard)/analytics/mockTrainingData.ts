// First, create mockTrainingData.ts in your project:
import type { TrainReportResponse } from '@/lib/backend';

export const mockTrainReport: TrainReportResponse = {
  before_train_metrics: {
    "PERSON": {
      precision: 0.75,
      recall: 0.68,
      fmeasure: 0.71
    },
    "ORGANIZATION": {
      precision: 0.82,
      recall: 0.77,
      fmeasure: 0.79
    }
  },
  after_train_metrics: {
    "PERSON": {
      precision: 0.85,
      recall: 0.79,
      fmeasure: 0.82
    },
    "ORGANIZATION": {
      precision: 0.88,
      recall: 0.84,
      fmeasure: 0.86
    }
  },
  after_train_examples: {
    true_positives: {
      "PERSON": [
        {
          source: "John Smith is the CEO",
          target: "B-PERSON I-PERSON O O O",
          index: 0
        },
        {
          source: "Meeting with Sarah Johnson tomorrow",
          target: "O O B-PERSON I-PERSON O",
          index: 1
        }
      ],
      "ORGANIZATION": [
        {
          source: "Microsoft announced new products",
          target: "B-ORGANIZATION O O O",
          index: 0
        },
        {
          source: "Working at Apple Inc",
          target: "O O B-ORGANIZATION I-ORGANIZATION",
          index: 1
        }
      ]
    },
    false_positives: {
      "PERSON": [
        {
          source: "The Washington Post reported",
          target: "O B-PERSON I-PERSON O",
          index: 0
        }
      ],
      "ORGANIZATION": [
        {
          source: "Summer Group meeting",
          target: "O B-ORGANIZATION O",
          index: 0
        }
      ]
    },
    false_negatives: {
      "PERSON": [
        {
          source: "Email from Dr. Williams",
          target: "O O O O",
          index: 0
        }
      ],
      "ORGANIZATION": [
        {
          source: "Report from Tesla today",
          target: "O O O O",
          index: 0
        }
      ]
    }
  }
};