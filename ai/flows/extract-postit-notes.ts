'use server';

/**
 * @fileOverview Extracts text, background color, and location from Post-it notes in an image.
 *
 * - extractPostitNotes - A function that handles the Post-it note extraction process.
 * - ExtractPostitNotesInput - The input type for the extractPostitNotes function.
 * - ExtractPostitNotesOutput - The return type for the extractPostitNotes function.
 */

import {ai} from '@/ai/ai-instance';
import {z} from 'genkit';

const ExtractPostitNotesInputSchema = z.object({
  photoDataUri: z
    .string()
    .describe(
      "A photo containing Post-it notes, as a data URI that must include a MIME type and use Base64 encoding. Expected format: 'data:<mimetype>;base64,<encoded_data>'."
    ),
});
export type ExtractPostitNotesInput = z.infer<typeof ExtractPostitNotesInputSchema>;

const ExtractPostitNotesOutputSchema = z.array(
  z.object({
    background_color: z.string().describe('The background color of the Post-it note (e.g., \'#FFFF00\').'),
    location: z
      .object({
        x: z.number().describe('The x-coordinate of the Post-it note in the image.'),
        y: z.number().describe('The y-coordinate of the Post-it note in the image.'),
      })
      .describe('The location of the Post-it note in the image.'),
    scale: z.number().describe('The scale of the Post-it note.'),
    size: z
      .object({
        height: z.number().describe('The height of the Post-it note.'),
        width: z.number().describe('The width of the Post-it note.'),
      })
      .describe('The size of the Post-it note.'),
    state: z.string().describe('The state of the Post-it note (e.g., \'normal\').'),
    text: z.string().describe('The text content of the Post-it note.'),
    widget_type: z.string().describe('The type of widget (always \'Note\' for Post-it notes).'),
  })
);
export type ExtractPostitNotesOutput = z.infer<typeof ExtractPostitNotesOutputSchema>;

export async function extractPostitNotes(input: ExtractPostitNotesInput): Promise<ExtractPostitNotesOutput> {
  return extractPostitNotesFlow(input);
}

const prompt = ai.definePrompt({
  name: 'extractPostitNotesPrompt',
  input: {
    schema: z.object({
      photoDataUri: z
        .string()
        .describe(
          "A photo containing Post-it notes, as a data URI that must include a MIME type and use Base64 encoding. Expected format: 'data:<mimetype>;base64,<encoded_data>'."
        ),
    }),
  },
  output: {
    schema: z.array(
      z.object({
        background_color: z.string().describe('The background color of the Post-it note (e.g., \'#FFFF00\').'),
        location: z
          .object({
            x: z.number().describe('The x-coordinate of the Post-it note in the image.'),
            y: z.number().describe('The y-coordinate of the Post-it note in the image.'),
          })
          .describe('The location of the Post-it note in the image.'),
        scale: z.number().describe('The scale of the Post-it note.'),
        size: z
          .object({
            height: z.number().describe('The height of the Post-it note.'),
            width: z.number().describe('The width of the Post-it note.'),
          })
          .describe('The size of the Post-it note.'),
        state: z.string().describe('The state of the Post-it note (e.g., \'normal\').'),
        text: z.string().describe('The text content of the Post-it note.'),
        widget_type: z.string().describe('The type of widget (always \'Note\' for Post-it notes).'),
      })
    ),
  },
  prompt: `You are an AI that can extract the content, background color and location of each of the post it notes from an image. You will return a JSON array containing objects representing each postit note.

Analyze the image and extract data for each post-it note:

Image: {{media url=photoDataUri}}

Return a JSON array of postit note objects. Each object should have the following structure:

{
  "background_color": "#FFFF00",
  "location": {
    "x": 1080,
    "y": 520
  },
  "scale": 1,
  "size": {
    "height": 200,
    "width": 200
  },
  "state": "normal",
  "text": "Bottom Right",
  "widget_type": "Note"
}
`,
});

const extractPostitNotesFlow = ai.defineFlow<
  typeof ExtractPostitNotesInputSchema,
  typeof ExtractPostitNotesOutputSchema
>(
  {
    name: 'extractPostitNotesFlow',
    inputSchema: ExtractPostitNotesInputSchema,
    outputSchema: ExtractPostitNotesOutputSchema,
  },
  async input => {
    const {output} = await prompt(input);
    return output!;
  }
);
