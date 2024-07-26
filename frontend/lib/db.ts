/**
 * @YeCao This is how the table was created:

 * CREATE TYPE model_type AS ENUM ('semantic search model', 'rag model', 'ner model');

 * CREATE TYPE status AS ENUM ('active', 'inactive', 'archived');

 * CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    image_url TEXT NOT NULL,
    name TEXT NOT NULL,
    status status NOT NULL,
    trained_at TIMESTAMP NOT NULL,
    description TEXT,
    deploy_endpoint_url TEXT,
    on_disk_size_kb NUMERIC(10, 2) NOT NULL,
    ram_size_kb NUMERIC(10, 2) NOT NULL,
    number_parameters INTEGER NOT NULL,
    rlhf_counts INTEGER NOT NULL,
    model_type model_type NOT NULL,
  );

  * INSERT INTO models (
      name, 
      image_url, 
      status, 
      trained_at, 
      description, 
      deploy_endpoint_url, 
      on_disk_size_kb, 
      ram_size_kb, 
      number_parameters, 
      rlhf_counts
    ) VALUES (
      'ThirdAI NER', 
      '/thirdai-small.png', 
      'active', 
      '2024-07-23T17:50:02.904Z', 
      'This is an NER model trained for showcase of these fields: ["O", "PHONENUMBER", "SSN", "CREDITCARDNUMBER", "LOCATION", "NAME"]', 
      'thirdai.com', 
      300 * 1024,  -- 300 MB converted to KB
      300 * 1024 * 2,  -- 300 * 2 MB converted to KB
      51203077, 
      0,
      'ner model'
    );
 */

import 'server-only';

import { neon } from '@neondatabase/serverless';
import { drizzle } from 'drizzle-orm/neon-http';
import {
  pgTable,
  text,
  numeric,
  integer,
  timestamp,
  pgEnum,
  serial
} from 'drizzle-orm/pg-core';
import { count, eq, ilike } from 'drizzle-orm';
import { createInsertSchema } from 'drizzle-zod';

export const db = drizzle(neon(process.env.POSTGRES_URL!));

export const statusEnum = pgEnum('status', ['active', 'inactive', 'archived']);
export const modelTypeEnum = pgEnum('model_type', ['semantic search model', 'rag model', 'ner model']);

export const models = pgTable('models', {
  id: serial('id').primaryKey(),
  imageUrl: text('image_url').notNull(),
  name: text('name').notNull(),
  status: statusEnum('status').notNull(),
  trainedAt: timestamp('trained_at').notNull(),
  description: text('description'),
  deployEndpointUrl: text('deploy_endpoint_url'),
  onDiskSizeKb: numeric('on_disk_size_kb', { precision: 10, scale: 2 }).notNull(),
  ramSizeKb: numeric('ram_size_kb', { precision: 10, scale: 2 }).notNull(),
  numberParameters: integer('number_parameters').notNull(),
  rlhfCounts: integer('rlhf_counts').notNull(),
  modelType: modelTypeEnum('model_type').notNull(),
});

export type SelectModel = typeof models.$inferSelect;
export const insertModelSchema = createInsertSchema(models);

export async function getModels(
  search: string,
  offset: number
): Promise<{
  models: SelectModel[];
  newOffset: number | null;
  totalModels: number;
}> {
  // Always search the full table, not per page
  if (search) {
    return {
      models: await db
        .select()
        .from(models)
        .where(ilike(models.name, `%${search}%`))
        .limit(1000),
      newOffset: null,
      totalModels: 0
    };
  }

  if (offset === null) {
    return { models: [], newOffset: null, totalModels: 0 };
  }

  let totalModels = await db.select({ count: count() }).from(models);
  let moreModels = await db.select().from(models).limit(5).offset(offset);
  let newOffset = moreModels.length >= 5 ? offset + 5 : null;

  return {
    models: moreModels,
    newOffset,
    totalModels: totalModels[0].count
  };
}

export async function deleteModelById(id: number) {
  await db.delete(models).where(eq(models.id, id));
}
