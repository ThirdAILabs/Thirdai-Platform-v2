import React, { useCallback, useContext, useEffect, useRef, useState } from 'react';
import { ModelService, Source } from '../modelServices';
import { ModelServiceContext } from '../Context';
import FilterAltOutlinedIcon from '@mui/icons-material/FilterAltOutlined';
import { Button } from '@mui/material'

const ConstraintSearch = () => {
    const [sources, setSources] = useState<Source[]>([]);
    const modelService = useContext<ModelService | null>(ModelServiceContext);

    const getSources = async () => {
        if (modelService) {
            const data = await modelService.sources();
            setSources(data);
        }
    }
    useEffect(() => {
        getSources();
    }, [modelService]);
    console.log("source: ", sources);
    return (
        <Button
            variant='contained'
        >
            <FilterAltOutlinedIcon fontSize='large' />
        </Button>
    )
}

export default ConstraintSearch;