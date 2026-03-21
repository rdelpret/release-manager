"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { useEffect } from "react";

interface RichTextEditorProps {
  content: Record<string, unknown> | null | undefined;
  onUpdate: (content: Record<string, unknown>) => void;
}

export function RichTextEditor({ content, onUpdate }: RichTextEditorProps) {
  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder: "Add a description..." }),
    ],
    content: content ?? undefined,
    editorProps: {
      attributes: {
        class:
          "prose prose-invert prose-sm max-w-none min-h-[100px] p-3 focus:outline-none text-text-primary [&_p]:my-1 [&_ul]:my-1 [&_ol]:my-1",
      },
    },
    onBlur: ({ editor }) => {
      onUpdate(editor.getJSON() as Record<string, unknown>);
    },
  });

  useEffect(() => {
    if (editor && content) {
      const current = JSON.stringify(editor.getJSON());
      const incoming = JSON.stringify(content);
      if (current !== incoming) {
        editor.commands.setContent(content);
      }
    }
  }, [content, editor]);

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <EditorContent editor={editor} />
    </div>
  );
}
