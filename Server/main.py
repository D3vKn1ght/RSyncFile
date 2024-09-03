from fastapi import FastAPI, File, UploadFile, HTTPException, Header
import os
import gzip
import shutil

app = FastAPI()

UPLOAD_FOLDER = "/upload"
os.makedirs(UPLOAD_FOLDER, exist_ok=True)

@app.post("/uploadfile/")
async def upload_file(file: UploadFile = File(...), filename: str = Header(None, alias="X-Filename")):
    if not filename:
        raise HTTPException(status_code=400, detail="Filename header is missing")
    
    try:
        print(f"Uploading file: {filename}")

        relative_path = filename.rstrip(".gz")
        save_path = os.path.normpath(os.path.join(UPLOAD_FOLDER, relative_path.lstrip("/")))

        if not save_path.startswith(UPLOAD_FOLDER):
            raise HTTPException(status_code=400, detail="Invalid file path")

        os.makedirs(os.path.dirname(save_path), exist_ok=True)

        tmp_file_location = f"{save_path}.tmp"
        content = await file.read()
        with open(tmp_file_location, "wb") as tmp_file:
            tmp_file.write(content)

        with gzip.open(tmp_file_location, 'rb') as f_in:
            with open(save_path, 'wb') as f_out:
                shutil.copyfileobj(f_in, f_out)

        os.remove(tmp_file_location)
        print(f"File uploaded and saved at {save_path}")
        return {"filename": save_path, "status": "File uploaded, extracted, and saved successfully"}

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"An error occurred during file upload: {str(e)}")


@app.delete("/deletefile/")
async def delete_file(filename: str = Header(...)):
    try:
        file_location = os.path.normpath(os.path.join(UPLOAD_FOLDER, filename.lstrip("/")))
        print(f"Deleting file: {file_location}")
        if not file_location.startswith(UPLOAD_FOLDER):
            raise HTTPException(status_code=400, detail="Invalid file path")

        if os.path.exists(file_location):
            os.remove(file_location)
            print(f"File {file_location} deleted successfully")
            return {"status": "File deleted successfully"}
        else:
            raise HTTPException(status_code=404, detail="File not found")
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"An error occurred during file deletion: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=7888)
