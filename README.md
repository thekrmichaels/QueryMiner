# QueryMiner

Extracts SQL from PHP source code and converts legacy queries into PostgreSQL-compatible parameterized statements ($1, $2, ...), normalizing query structure.

## Installation

### Steps to Run

1. Download the corresponding archive for your operating system.
2. Extract the contents of the archive.
3. Navigate to the folder where the executable file is located.

## Usage

Once extracted, you can run QueryMiner by accessing the appropriate executable from the command line.

- **Windows**:
  - Open Command Prompt or PowerShell and navigate to the folder where `queryminer.exe` is located.
  - Run the command:

    ```sh
    .\queryminer.exe <source_path> <destination_folder>
    ```

- **Linux / macOS**:
  - Open a terminal and navigate to the folder where `queryminer` is located.
  - Run the command:

    ```sh
    ./queryminer <source_path> <destination_folder>
    ```

Replace `<source_path>` with the path to your PHP files and `<destination_folder>` with the folder where you want to save the extracted queries.
