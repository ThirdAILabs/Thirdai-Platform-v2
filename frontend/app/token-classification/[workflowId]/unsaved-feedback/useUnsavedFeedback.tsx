'use client';

import { useParams } from 'next/navigation';
import { useEffect, useMemo, useState } from 'react';

export interface TaggedToken {
  token: string;
  tag: string;
}

interface UnsavedFeedback {
  prediction: TaggedToken[];
  annotation: TaggedToken[];
}

export interface UpdatedEntity {
  text: string;
  tag: string;
  remove: () => void;
}

interface UseUnsavedFeedback {
  prediction: TaggedToken[];
  annotation: TaggedToken[];
  handlePrediction: (prediction: TaggedToken[]) => void;
  handleAnnotation: (annotation: TaggedToken[]) => void;
  updatedEntities: UpdatedEntity[];
}

const thereIsFeedback = (
  prediction: TaggedToken[],
  annotation: TaggedToken[]
) => {
  for (let i = 0; i < prediction.length; i++) {
    if (prediction[i].tag != annotation[i].tag) {
      return true;
    }
  }
  return false;
};

function useUnsavedFeedbackCache() {
  const params = useParams();
  const [cache, setCache] = useState<UnsavedFeedback[]>([]);

  const localStorageName = useMemo(
    () => `unsaved-feedback-${params.workflowId as string}`,
    []
  );

  useEffect(() => {
    const feedbackListJsonString = localStorage.getItem(localStorageName);
    if (feedbackListJsonString) {
      setCache(JSON.parse(feedbackListJsonString));
    }
  }, []);

  const saveToLocalStorage = (newCache: UnsavedFeedback[]) => {
    localStorage.setItem(localStorageName, JSON.stringify(newCache));
  };

  const add = (feedback: UnsavedFeedback) => {
    const newCache = [...cache, feedback];
    saveToLocalStorage(newCache);
    setCache(newCache);
  };

  const updateFeedback = (index: number, newAnnotation: TaggedToken[]) => {
    const newCache = cache
      .map((feedback, feedbackIndex) => {
        if (feedbackIndex !== index) {
          return feedback;
        }
        if (thereIsFeedback(feedback.prediction, newAnnotation)) {
          return {
            prediction: feedback.prediction,
            annotation: newAnnotation
          };
        }
        return { prediction: [], annotation: [] };
      })
      .filter(
        (feedback) =>
          feedback.prediction.length > 0 && feedback.annotation.length > 0
      );
    saveToLocalStorage(newCache);
    setCache(newCache);
  };

  return {
    cache,
    add,
    updateFeedback,
    save: (newFeedback?: UnsavedFeedback) => {
      let newCache = cache;
      if (newFeedback) {
        newCache = [...newCache, newFeedback];
      }
      saveToLocalStorage(newCache);
    }
  };
}

class UpdatedEntityBuilders {
  builders: { positions: number[], tag: string }[];

  constructor() {
    this.builders = [];
  }

  add(position: number, tag: string) {
    if (
      this.notEmpty() &&
      tag === this.lastBuilder().tag &&
      position === this.lastBuilderLastPosition() + 1
    ) {
      this.lastBuilder().positions.push(position);
    } else {
      this.builders.push({ positions: [position], tag });
    }
  }

  private notEmpty() {
    return this.builders.length > 0;
  }

  private lastBuilder() {
    return this.builders[this.builders.length - 1];
  }

  private lastBuilderLastPosition() {
    const lastBuilder = this.lastBuilder();
    return lastBuilder.positions[lastBuilder.positions.length - 1];
  }
}

const getUpdatedEntities = (prediction: TaggedToken[], annotation: TaggedToken[], setAnnotation: (annotation: TaggedToken[]) => void) => {
  const builders = new UpdatedEntityBuilders();
    annotation.forEach((anno, pos) => {
      if (anno.tag !== prediction[pos].tag) {
        builders.add(pos, anno.tag);
      }
    });
    return builders.builders.map(({ positions, tag }) => {
      const revertedAnnotation = annotation.map(({ token, tag }, index) => ({
        token,
        tag: positions.includes(index) ? prediction[index].tag : tag
      }));
      return {
        text: positions.map((pos) => annotation[pos].token).join(' '),
        tag,
        remove: () => setAnnotation(revertedAnnotation)
      };
    });
};

export default function useUnsavedFeedback(): UseUnsavedFeedback {
  const cache = useUnsavedFeedbackCache();
  const [prediction, setPrediction] = useState<TaggedToken[]>([]);
  const [annotation, setAnnotation] = useState<TaggedToken[]>([]);

  const updatedEntities = useMemo((): UpdatedEntity[] => {
    let updatedEntities = cache.cache.map(({prediction, annotation}, index) => {
      return getUpdatedEntities(prediction, annotation, (annotation) => cache.updateFeedback(index, annotation));
    });
    updatedEntities.push(getUpdatedEntities(prediction, annotation, setAnnotation));
    return updatedEntities.flatMap(x => x);
  }, [prediction, annotation, cache.cache]);

  const handlePrediction = (newPrediction: TaggedToken[]) => {
    if (thereIsFeedback(prediction, annotation)) {
      cache.add({ prediction, annotation });
    }
    setPrediction(newPrediction);
    setAnnotation(newPrediction);
  };

  const handleAnnotation = (newAnnotation: TaggedToken[]) => {
    if (thereIsFeedback(prediction, newAnnotation)) {
      cache.save(/* newFeedback= */ { prediction, annotation: newAnnotation });
    }
    cache.save();
  };

  return {
    prediction,
    annotation,
    handlePrediction,
    handleAnnotation,
    updatedEntities
  };
}
