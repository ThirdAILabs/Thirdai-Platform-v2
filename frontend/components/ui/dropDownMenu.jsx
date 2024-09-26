// import React, { useState } from 'react';

// const Dropdown = ({ title, handleSelectedTeam, teams }) => {
//   const [menuTitle, setMenuTitle] = useState(title);
//   const handleChange = (e) => {
//     setMenuTitle(e.target.value);
//     handleSelectedTeam(e.target.value);
//   }
//   return (
//     <select
//       value={menuTitle}
//       onChange={handleChange}
//       className="border border-gray-300 rounded-2xl px-4 py-2 shadow-md"
//     >
//       <option value="">{title}</option>
//       {teams.map((team) => (
//         <option key={team.id} value={team.name}>
//           {team.name}
//         </option>
//       ))}
//     </select>
//   );
// };

// export default Dropdown;


import React, { useState } from 'react';

const Dropdown = ({ title, handleSelectedTeam, teams }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [menuTitle, setMenuTitle] = useState(title);

  const handleToggle = () => {
    setIsOpen(!isOpen);
  };

  const handleChange = (team) => {
    setMenuTitle(team.name);
    handleSelectedTeam(team.name);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <button
        onClick={handleToggle}
        className="border border-gray-300 rounded-2xl bg-white px-4 py-2 shadow-md flex justify-between items-center w-full"
      >
        {menuTitle}
        <span className="ml-2">{isOpen ? '▲' : '▼'}</span>
      </button>
      {isOpen && (
        <div className="absolute left-0 right-0 mt-1 bg-white rounded-2xl border border-gray-500 shadow-lg z-10">
          {teams.map((team) => (
            <div
              key={team.id}
              onClick={() => handleChange(team)}
              className="px-4 py-2 hover:bg-blue-500 cursor-pointer rounded-2xl"
            >
              {team.name}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Dropdown;

