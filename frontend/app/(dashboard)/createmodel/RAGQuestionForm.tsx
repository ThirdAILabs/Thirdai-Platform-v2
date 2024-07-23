import React from 'react';

const RAGQuestionForm = () => {
//   const handleFileUpload = (event) => {
//     const files = event.target.files;
//     // Handle file processing here
//     console.log(files);
//   };

  return (
    <div>
      <h2>RAG Model Questions</h2>
      <p>Please upload the necessary files for the RAG model.</p>
      <input type="file" multiple onChange={()=>{}} />
    </div>
  );
};

export default RAGQuestionForm;
