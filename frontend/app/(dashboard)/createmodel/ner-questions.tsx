// app/NERQuestions.js

import React, { useState } from 'react';

type Category = {
  name: string;
  example: string;
  description: string;
};

const predefinedChoices = [
  'PHONENUMBER',
  'SSN',
  'CREDITCARDNUMBER',
  "LOCATION",
  "NAME"
];

const NERQuestions = () => {
  const [categories, setCategories] = useState([{ name: '', example: '', description: '' }]);

  const handleCategoryChange = (index: number, field: keyof Category, value: string) => {
    const updatedCategories = [...categories];
    updatedCategories[index][field] = value;
    setCategories(updatedCategories);
  };

  const handleAddCategory = () => {
    setCategories([...categories, { name: '', example: '', description: '' }]);
  };

  const handleRemoveCategory = (index: number) => {
    const updatedCategories = categories.filter((_, i) => i !== index);
    setCategories(updatedCategories);
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log('Categories:', categories);
    // Handle form submission logic here
  };

  return (
    <div className='p-5'>
      <h3 className='mb-3 text-lg font-semibold'>Specify Tokens</h3>
      <form onSubmit={handleSubmit}>
        {categories.map((category, index) => (
          <div key={index} className='flex flex-col md:flex-row md:items-center my-2'>
            <div className="relative w-full md:w-1/3">
              <input
                type="text"
                list={`category-options-${index}`}
                className="form-input w-full px-3 py-2 border rounded-md"
                placeholder="Category Name"
                value={category.name}
                onChange={(e) => handleCategoryChange(index, 'name', e.target.value)}
              />
              <datalist id={`category-options-${index}`}>
                {predefinedChoices.map((choice, i) => (
                  <option key={i} value={choice} />
                ))}
              </datalist>
            </div>
            <input
              type="text"
              className='form-input w-full md:w-1/3 md:ml-2 mt-2 md:mt-0 px-3 py-2 border rounded-md'
              placeholder="Example"
              value={category.example}
              onChange={(e) => handleCategoryChange(index, 'example', e.target.value)}
            />
            <input
              type="text"
              className='form-input w-full md:w-1/3 md:ml-2 mt-2 md:mt-0 px-3 py-2 border rounded-md'
              placeholder="What this category is about."
              value={category.description}
              onChange={(e) => handleCategoryChange(index, 'description', e.target.value)}
            />
            <button type="button" className='bg-red-500 text-white px-4 py-2 rounded-md md:ml-2 mt-2 md:mt-0' onClick={() => handleRemoveCategory(index)}>
              Remove
            </button>
          </div>
        ))}
        <button type="button" className='bg-blue-500 text-white px-4 py-2 rounded-md mt-2 mr-2' onClick={handleAddCategory}>
          Add Category
        </button>
        <button type="submit" className='bg-green-500 text-white px-4 py-2 rounded-md mt-2'>Finish and Review</button>
      </form>
    </div>
  );
};

export default NERQuestions;
