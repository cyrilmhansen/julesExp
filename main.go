package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Page names
const (
	menuPageName           = "menu"
	addOperationPageName   = "addOperation"
	listOperationsPageName = "listOperations"
	loadDataPageName       = "loadData"
	saveDataPageName       = "saveData"
	modalPageNamePrefix    = "modal_"
)

// Global variables
var (
	currentAccountData AccountData
	appInstance               *tview.Application
	pagesInstance             *tview.Pages
	globalAddOperationForm    *tview.Form
	globalListOperationsTable *tview.Table
	globalLoadDataForm        *tview.Form
	globalSaveDataForm        *tview.Form // New global for the save data form
	shouldShowAddFormAfterList bool       = false
)

// Operation struct defines the structure for an operation
type Operation struct {
	Date        time.Time
	Description string
	Debit       float64
	Credit      float64
}

// AccountData struct defines the structure for account data
type AccountData struct {
	Operations   []Operation
	LastSaveDate string // Store as "jj/mm/aa" string
}

const dateLayout = "02/01/06"

// sortOperations sorts a slice of Operation structs by date.
func sortOperations(ops []Operation) {
	sort.SliceStable(ops, func(i, j int) bool {
		return ops[i].Date.Before(ops[j].Date)
	})
}

// showModal displays a modal dialog with a message. Uses global pagesInstance.
func showModal(pageNameSuffix, title, message string, buttons []string, doneFunc func(buttonIndex int, buttonLabel string)) {
	modalName := modalPageNamePrefix + pageNameSuffix
	modal := tview.NewModal().
		SetText(message).
		SetTitle(title).
		AddButtons(buttons).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pagesInstance.RemovePage(modalName) // Remove the modal page itself
			if doneFunc != nil {
				doneFunc(buttonIndex, buttonLabel)
			}
		})
	pagesInstance.AddPage(modalName, modal, false, true)
	pagesInstance.ShowPage(modalName) // Show the modal
}

// clearFormFields resets the input fields of the provided form. Uses global appInstance for focus.
func clearFormFields(form *tview.Form) {
	if form == nil {
		return
	}
	// Iterate over form items and clear them
	for i := 0; i < form.GetFormItemCount(); i++ {
		item := form.GetFormItem(i)
		if inputField, ok := item.(*tview.InputField); ok {
			inputField.SetText("")
		}
	}
	// Only set focus if the form has items and the app instance is available
	if form.GetFormItemCount() > 0 && appInstance != nil && form.GetFormItem(0) != nil {
		appInstance.SetFocus(form.GetFormItem(0))
	}
}

// createAddOperationPage initializes the global globalAddOperationForm.
func createAddOperationPage() {
	globalAddOperationForm = tview.NewForm()

	dateField := tview.NewInputField().SetLabel("Date (JJ/MM/AA)").SetFieldWidth(10)
	descriptionField := tview.NewInputField().SetLabel("Description").SetFieldWidth(30)
	debitField := tview.NewInputField().SetLabel("Debit").SetFieldWidth(15)
	creditField := tview.NewInputField().SetLabel("Credit").SetFieldWidth(15)

	globalAddOperationForm.AddFormItem(dateField)
	globalAddOperationForm.AddFormItem(descriptionField)
	globalAddOperationForm.AddFormItem(debitField)
	globalAddOperationForm.AddFormItem(creditField)

	globalAddOperationForm.AddButton("Sauvegarder", func() {
		dateStr := strings.TrimSpace(dateField.GetText())
		descriptionStr := strings.TrimSpace(descriptionField.GetText())
		debitStr := strings.TrimSpace(debitField.GetText())
		creditStr := strings.TrimSpace(creditField.GetText())

		// Validate Description
		if descriptionStr == "" {
			showModal("descErr", "Erreur de validation", "La description ne peut pas être vide.", []string{"OK"}, nil)
			return
		}

		// Validate Date
		parsedDate, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			showModal("dateErr", "Erreur de validation", "Format de date invalide. Utilisez JJ/MM/AA.", []string{"OK"}, nil)
			return
		}

		// Validate Debit/Credit
		var debitVal float64
		var creditVal float64
		var errDebit, errCredit error

		if debitStr != "" {
			debitVal, errDebit = strconv.ParseFloat(debitStr, 64)
			if errDebit != nil || debitVal < 0 {
				showModal("debitErr", "Erreur de validation", "Valeur de débit invalide ou négative.", []string{"OK"}, nil)
				return
			}
		}

		if creditStr != "" {
			creditVal, errCredit = strconv.ParseFloat(creditStr, 64)
			if errCredit != nil || creditVal < 0 {
				showModal("creditErr", "Erreur de validation", "Valeur de crédit invalide ou négative.", []string{"OK"}, nil)
				return
			}
		}

		if debitStr != "" && creditStr != "" {
			showModal("debitCreditComboErr", "Erreur de validation", "Veuillez entrer un débit OU un crédit, pas les deux.", []string{"OK"}, nil)
			return
		}
		if debitStr == "" && creditStr == "" {
			showModal("debitCreditMissingErr", "Erreur de validation", "Veuillez entrer un débit ou un crédit.", []string{"OK"}, nil)
			return
		}

		// If validation passes
		newOperation := Operation{
			Date:        parsedDate,
			Description: descriptionStr,
			Debit:       debitVal,
			Credit:      creditVal,
		}
		currentAccountData.Operations = append(currentAccountData.Operations, newOperation)
		sortOperations(currentAccountData.Operations)

		showModal("saveSuccess", "Succès", "Opération sauvegardée !", []string{"OK"}, func(buttonIndex int, buttonLabel string) {
			clearFormFields(globalAddOperationForm) // Uses global form
		})
	})

	globalAddOperationForm.AddButton("Annuler", func() {
		clearFormFields(globalAddOperationForm)
		pagesInstance.SwitchToPage(menuPageName)
	})

	globalAddOperationForm.SetBorder(true).SetTitle("Ajouter une opération").SetCancelFunc(func() {
		clearFormFields(globalAddOperationForm) // Clear before switching
		pagesInstance.SwitchToPage(menuPageName)
	})
}

// createLoadDataPage initializes the global globalLoadDataForm.
func createLoadDataPage() {
	globalLoadDataForm = tview.NewForm()
	fileNameField := tview.NewInputField().SetLabel("Nom du fichier:").SetFieldWidth(40)

	globalLoadDataForm.AddFormItem(fileNameField)

	globalLoadDataForm.AddButton("Charger", func() {
		fileName := strings.TrimSpace(fileNameField.GetText())
		if fileName == "" {
			showModal("loadFileEmptyErr", "Erreur", "Veuillez entrer un nom de fichier.", []string{"OK"}, nil)
			return
		}

		confirmModalName := "confirmLoadFile"
		showModal(confirmModalName, "Confirmer le chargement",
			"Charger les données depuis '"+fileName+"' ? Les données non sauvegardées seront perdues.",
			[]string{"Oui", "Non"}, func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Oui" {
					loadedOps, err := loadData(fileName) // loadData is from fileops.go
					if err != nil {
						errorMsg := "Erreur lors du chargement du fichier: " + err.Error()
						if os.IsNotExist(err) {
							errorMsg = "Fichier '" + fileName + "' non trouvé."
						}
						showModal("loadFileErr", "Erreur Chargement", errorMsg, []string{"OK"}, nil)
						return
					}
					// Successfully loaded data
					currentAccountData = loadedOps // Update global account data
					sortOperations(currentAccountData.Operations)
					showModal("loadFileSuccess", "Succès", "Données chargées avec succès depuis '"+fileName+"'.", []string{"OK"},
						func(bi int, bl string) {
							pagesInstance.SwitchToPage(menuPageName)
						})
				}
			})
	})

	globalLoadDataForm.AddButton("Annuler", func() {
		fileNameField.SetText("") // Clear field on cancel
		pagesInstance.SwitchToPage(menuPageName)
	})

	globalLoadDataForm.SetBorder(true).SetTitle("Charger les Comptes").SetCancelFunc(func() {
		fileNameField.SetText("") // Clear field on Esc
		pagesInstance.SwitchToPage(menuPageName)
	})
}

// createSaveDataPage initializes the global globalSaveDataForm.
func createSaveDataPage() {
	globalSaveDataForm = tview.NewForm()
	saveDateField := tview.NewInputField().SetLabel("Date d'aujourd'hui (JJ/MM/AA):").SetFieldWidth(10)
	saveFileNameField := tview.NewInputField().SetLabel("Nom du fichier:").SetFieldWidth(40)

	globalSaveDataForm.AddFormItem(saveDateField)
	globalSaveDataForm.AddFormItem(saveFileNameField)

	globalSaveDataForm.AddButton("Sauvegarder", func() {
		saveDateStr := strings.TrimSpace(saveDateField.GetText())
		fileNameStr := strings.TrimSpace(saveFileNameField.GetText())

		if saveDateStr == "" {
			showModal("saveDateEmptyErr", "Erreur", "Veuillez entrer la date d'aujourd'hui.", []string{"OK"}, nil)
			return
		}
		_, err := time.Parse(dateLayout, saveDateStr)
		if err != nil {
			showModal("saveDateInvalidErr", "Erreur", "Format de date invalide pour la date d'aujourd'hui. Utilisez JJ/MM/AA.", []string{"OK"}, nil)
			return
		}
		if fileNameStr == "" {
			showModal("saveFileEmptyErr", "Erreur", "Veuillez entrer un nom de fichier.", []string{"OK"}, nil)
			return
		}

		confirmModalName := "confirmSaveFile"
		showModal(confirmModalName, "Confirmer la sauvegarde",
			"Sauvegarder les données dans '"+fileNameStr+"' ?",
			[]string{"Oui", "Non"}, func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Oui" {
					currentAccountData.LastSaveDate = saveDateStr
					err := saveData(currentAccountData, fileNameStr) // saveData is from fileops.go
					if err != nil {
						showModal("saveFileErr", "Erreur Sauvegarde", "Erreur lors de la sauvegarde du fichier: "+err.Error(), []string{"OK"}, nil)
						return
					}
					showModal("saveFileSuccess", "Succès", "Données sauvegardées avec succès dans '"+fileNameStr+"'.", []string{"OK"},
						func(bi int, bl string) {
							pagesInstance.SwitchToPage(menuPageName)
						})
				}
			})
	})

	globalSaveDataForm.AddButton("Annuler", func() {
		saveDateField.SetText(time.Now().Format(dateLayout)) // Reset to current date
		saveFileNameField.SetText("")
		pagesInstance.SwitchToPage(menuPageName)
	})

	globalSaveDataForm.SetBorder(true).SetTitle("Sauvegarder les Comptes").SetCancelFunc(func() {
		saveDateField.SetText(time.Now().Format(dateLayout)) // Reset to current date
		saveFileNameField.SetText("")
		pagesInstance.SwitchToPage(menuPageName)
	})
}

func main() {
	appInstance = tview.NewApplication()
	pagesInstance = tview.NewPages()

	currentAccountData = AccountData{
		Operations:   []Operation{},
		LastSaveDate: "",
	}

	// Initialize global page widgets
	createAddOperationPage()
	pagesInstance.AddPage(addOperationPageName, globalAddOperationForm, true, false)

	createListOperationsPage()
	pagesInstance.AddPage(listOperationsPageName, globalListOperationsTable, true, false)

	createLoadDataPage()
	pagesInstance.AddPage(loadDataPageName, globalLoadDataForm, true, false)

	createSaveDataPage() // Create the save data page
	pagesInstance.AddPage(saveDataPageName, globalSaveDataForm, true, false)

	menuList := tview.NewList().
		AddItem("Entree des operations", "Ajouter une nouvelle opération", '1', func() {
			showModal("confirmListBeforeAdd", "Afficher Liste?", "Voulez-vous afficher la liste des opérations existantes avant la saisie ?", []string{"Oui", "Non"},
				func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Oui" {
						shouldShowAddFormAfterList = true
						refreshAndShowListOperationsPage()
					} else { // "Non"
						clearFormFields(globalAddOperationForm)
						pagesInstance.SwitchToPage(addOperationPageName)
					}
				})
		}).
		AddItem("Liste des operations", "Afficher toutes les opérations", '2', func() {
			refreshAndShowListOperationsPage()
		}).
		AddItem("Chargement des comptes", "Charger les données d'un fichier", '3', func() {
			if globalLoadDataForm != nil {
				// Assuming the first item is the filename field for simplicity here
				if item := globalLoadDataForm.GetFormItem(0); item != nil {
					if inputField, ok := item.(*tview.InputField); ok {
						inputField.SetText("") // Clear filename
					}
				}
				if globalLoadDataForm.GetFormItemCount() > 0 {
					appInstance.SetFocus(globalLoadDataForm.GetFormItem(0))
				}
			}
			pagesInstance.SwitchToPage(loadDataPageName)
		}).
		AddItem("Sauvegarde des comptes", "Sauvegarder les données dans un fichier", '4', func() {
			if globalSaveDataForm != nil {
				if dateItem := globalSaveDataForm.GetFormItemByLabel("Date d'aujourd'hui (JJ/MM/AA):"); dateItem != nil {
					if dateInputField, ok := dateItem.(*tview.InputField); ok {
						dateInputField.SetText(time.Now().Format(dateLayout))
					}
				}
				if fileItem := globalSaveDataForm.GetFormItemByLabel("Nom du fichier:"); fileItem != nil {
					if fileInputField, ok := fileItem.(*tview.InputField); ok {
						fileInputField.SetText("")
					}
				}
				if globalSaveDataForm.GetFormItemCount() > 1 {
					appInstance.SetFocus(globalSaveDataForm.GetFormItem(1))
				} else if globalSaveDataForm.GetFormItemCount() > 0 {
					appInstance.SetFocus(globalSaveDataForm.GetFormItem(0))
				}
			}
			pagesInstance.SwitchToPage(saveDataPageName)
		}).
		AddItem("Effacement de ligne", "Supprimer une opération de la liste", '5', func() { // Updated description
			refreshAndShowListOperationsPage()
			// Set focus to the table to make selection and deletion easier
			if appInstance != nil && globalListOperationsTable != nil { // Ensure not nil before focusing
				appInstance.SetFocus(globalListOperationsTable)
			}
		}).
		AddItem("Position actuel du compte", "Afficher le solde actuel", '6', nil).
		AddItem("Fin des operations", "Quitter l'application", '7', func() {
			appInstance.Stop()
		})
	menuList.SetBorder(true).SetTitle("Menu Principal")

	pagesInstance.AddPage(menuPageName, menuList, true, true)
	appInstance.SetRoot(pagesInstance, true).EnableMouse(true)

	if err := appInstance.Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
		panic(err)
	}
}

// populateTableData fills the globalListOperationsTable with operations.
func populateTableData() {
	if globalListOperationsTable == nil { return }
	globalListOperationsTable.Clear() // Clear existing data first
	headers := []string{"Date", "Description", "Débit", "Crédit", "Solde"}
	headerStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorYellow)

	for c, header := range headers {
		globalListOperationsTable.SetCell(0, c, tview.NewTableCell(header).SetStyle(headerStyle).SetSelectable(false))
	}

	var runningBalance float64
	for r, op := range currentAccountData.Operations {
		globalListOperationsTable.SetCell(r+1, 0, tview.NewTableCell(op.Date.Format(dateLayout)).SetAlign(tview.AlignRight))
		globalListOperationsTable.SetCell(r+1, 1, tview.NewTableCell(op.Description).SetAlign(tview.AlignLeft))
		globalListOperationsTable.SetCell(r+1, 2, tview.NewTableCell(fmt.Sprintf("%.2f", op.Debit)).SetAlign(tview.AlignRight))
		globalListOperationsTable.SetCell(r+1, 3, tview.NewTableCell(fmt.Sprintf("%.2f", op.Credit)).SetAlign(tview.AlignRight))
		runningBalance += op.Credit - op.Debit
		globalListOperationsTable.SetCell(r+1, 4, tview.NewTableCell(fmt.Sprintf("%.2f", runningBalance)).SetAlign(tview.AlignRight))
	}
	globalListOperationsTable.ScrollToBeginning()
}

// createListOperationsPage initializes the global globalListOperationsTable.
func createListOperationsPage() {
	globalListOperationsTable = tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false) // Ensure selectable is true for row deletion

	globalListOperationsTable.SetTitle("Liste des Opérations").SetBorder(true)
	globalListOperationsTable.SetFixed(1, 0)

	populateTableData() // Initial population

	globalListOperationsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if shouldShowAddFormAfterList { // Specific state check first
				shouldShowAddFormAfterList = false // Reset flag
				clearFormFields(globalAddOperationForm)
				pagesInstance.SwitchToPage(addOperationPageName)
				return nil // Event handled
			}
			// Default Escape action
			pagesInstance.SwitchToPage(menuPageName)
			return nil // Event handled
		}

		// Delete Operation Logic (Del key or 'd'/'D')
		if event.Key() == tcell.KeyDelete || (event.Key() == tcell.KeyRune && (event.Rune() == 'd' || event.Rune() == 'D')) {
			selectedRow, _ := globalListOperationsTable.GetSelection()
			rowIndexInSlice := selectedRow - 1 // Adjust for header row

			// Validate selection: cannot delete header or if selection is invalid/out of bounds
			if selectedRow == 0 {
				showModal("delHeaderErrModal", "Action Impossible", "La ligne d'en-tête ne peut pas être supprimée.", []string{"OK"}, nil)
				return nil // Event handled by showing modal
			}
			if rowIndexInSlice < 0 || rowIndexInSlice >= len(currentAccountData.Operations) {
				showModal("delInvalidRowErrModal", "Action Impossible", "Veuillez sélectionner une ligne d'opération valide à supprimer.", []string{"OK"}, nil)
				return nil // Event handled by showing modal
			}

			opToDelete := currentAccountData.Operations[rowIndexInSlice]
			confirmMessage := fmt.Sprintf("Supprimer l'opération suivante ?\n\nDate: %s\nDescription: %s\nDébit: %.2f\nCrédit: %.2f",
				opToDelete.Date.Format(dateLayout), opToDelete.Description, opToDelete.Debit, opToDelete.Credit)

			showModal("confirmDeleteOpModal", "Confirmer Suppression", confirmMessage, []string{"Oui", "Non"},
				func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Oui" {
						currentAccountData.Operations = append(currentAccountData.Operations[:rowIndexInSlice], currentAccountData.Operations[rowIndexInSlice+1:]...)
						populateTableData() // Refresh the table content
						showModal("deleteSuccessModal", "Succès", "Opération supprimée.", []string{"OK"}, nil)

						// Adjust selection after deletion for better UX
						if len(currentAccountData.Operations) == 0 {
							globalListOperationsTable.Select(0, 0) // Select header if table is empty
						} else if selectedRow > len(currentAccountData.Operations) {
							// If last element was deleted (selectedRow is 1-based index including header)
							globalListOperationsTable.Select(len(currentAccountData.Operations), 0) // Select new last element (which is now at the previous last data row index)
						} else {
							// Try to select the same row index, which now contains the next item, or the new last item
							globalListOperationsTable.Select(selectedRow, 0)
						}
						if appInstance != nil { // Ensure appInstance is not nil before using
							appInstance.SetFocus(globalListOperationsTable) // Keep focus on table
						}
					}
				})
			return nil // Event handled by initiating deletion process or doing nothing
		}
		return event // Event not handled
	})
}

// refreshAndShowListOperationsPage clears, repopulates, and switches to the list operations page.
// Uses global pagesInstance and globalListOperationsTable.
func refreshAndShowListOperationsPage() {
	if globalListOperationsTable == nil { // Should not happen if initialized in main
		createListOperationsPage()
		pagesInstance.AddPage(listOperationsPageName, globalListOperationsTable, true, false)
	}
	populateTableData() // This now clears and repopulates
	pagesInstance.SwitchToPage(listOperationsPageName)
}
