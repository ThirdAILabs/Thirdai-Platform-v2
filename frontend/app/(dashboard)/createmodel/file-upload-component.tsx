import React, { ChangeEvent, useRef, useState } from 'react';

interface FileUploadProps {
  index: number;
  setSourceValue: (index: number, files: FileList | null) => void;
  allowedFileTypes?: string[];
}

const FileUpload: React.FC<FileUploadProps> = ({
  index,
  setSourceValue,
  allowedFileTypes = ['csv', 'pdf', 'docx'],
}) => {
  const [fileCount, setFileCount] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  // Check if file type is allowed
  const isAllowedFileType = (fileName: string, allowedTypes: string[]): boolean => {
    const fileExt = '' + fileName.split('.').pop()?.toLowerCase();
    return allowedTypes.includes(fileExt);
  };

  const handleUpload = (e: ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;

    if (!files) {
      setSourceValue(index, null);
      setFileCount(0);
      return;
    }

    // Convert FileList to Array and filter
    const filesArray = Array.from(files);
    const filteredFiles = filesArray.filter((file) =>
      isAllowedFileType(file.name, allowedFileTypes)
    );

    // Create a new FileList from the filtered files
    const filteredFileList = new DataTransfer();
    filteredFiles.forEach((file) => filteredFileList.items.add(file));

    // Update file count
    setFileCount(filteredFiles.length);

    // Call setSourceValue with the filtered FileList
    setSourceValue(index, filteredFileList.files);
  };

  const openFilePicker = () => {
    if (inputRef.current) {
      inputRef.current.value = ''; // Clear previous selection
      inputRef.current.click();
    }
  };

  const openFolderPicker = () => {
    if (inputRef.current) {
      inputRef.current.value = ''; // Clear previous selection
      inputRef.current.setAttribute('webkitdirectory', 'true');
      inputRef.current.click();
    }
  };

  return (
    <div className="flex flex-col space-y-2">
      <div className="flex space-x-2 items-center">
        <button
          type="button"
          className="text-md bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          onClick={() => {
            if (inputRef.current) {
              inputRef.current.removeAttribute('webkitdirectory');
              openFilePicker();
            }
          }}
        >
          Choose Files
        </button>
        <button
          type="button"
          className="text-md bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
          onClick={openFolderPicker}
        >
          Choose Folder
        </button>
        {fileCount > 0 && (
          <span className="text-md text-gray-700">
            {fileCount} file{fileCount !== 1 ? 's' : ''} selected
          </span>
        )}
      </div>
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        onChange={handleUpload}
        accept={allowedFileTypes.join(',')}
        multiple
      />
    </div>
  );
};

export default FileUpload;
